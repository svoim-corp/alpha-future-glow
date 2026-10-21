# Origination — заявка

Зона Origination (`backend/internal/origination`, разработчик A): черновик, этап 1 и этап 2, проверка полноты,
отправка, машина состояний и история. Схема БД — `origination`. Статус заявки меняет только эта зона
(`shared.Applications.Transition`).

Данные компании, обороты, полномочия и кандидатов в поручители клиент не вводит: их даёт зона Company по данным банка.
От клиента нужны только решения и согласия: продукт, сумма, срок, цель, согласия на БКИ, выбор поручителя,
подтверждение достоверности.

## Доступ

Все клиентские эндпоинты — для пользователя с компанией (`CLIENT`). Чужая заявка — `403 FORBIDDEN`, сотрудник банка
(без компании) — `403 FORBIDDEN`. Удалённый черновик для клиента не существует — `404 NOT_FOUND`.

## Машина состояний

Реализовано (`internal/origination/machine.go`). Статус меняется только через `Transition` (для других зон —
`shared.Applications.Transition`). Переход выполняется в транзакции под блокировкой строки заявки: из одновременных
переходов из одного статуса проходит один, остальные получают `409`. Вместе со статусом увеличивается `version`
и пишется событие истории (`StatusEvent`: из какого статуса, в какой, действие, кто, комментарий, время). После
коммита публикуется `ApplicationStatusChanged` (его слушают очередь банка и конвейер Scoring); сбой подписчика
логируется и не отменяет переход.

| Из | В | Кто | Действие (`action`) |
|---|---|---|---|
| `DRAFT` | `SUBMITTED` | клиент | `SUBMIT` |
| `DRAFT` | `CLOSED` | клиент удалил черновик | `DELETE_DRAFT` |
| `SUBMITTED` | `DATA_CHECK` | конвейер Scoring | `CHECKS_DONE` |
| `DATA_CHECK` | `SCORING` | конвейер Scoring | `SCORING_STARTED` |
| `DATA_CHECK`, `ON_DECISION` | `INFO_REQUESTED` | сотрудник | `INFO_REQUEST` |
| `INFO_REQUESTED` | `DATA_CHECK` | клиент ответил | `INFO_ANSWERED` |
| `SCORING` | `ON_DECISION` | конвейер Scoring | `SCORED` |
| `ON_DECISION` | `APPROVED`, `COUNTER_OFFER`, `REJECTED` | авторешение или сотрудник | `DECISION` |
| `APPROVED`, `COUNTER_OFFER` | `SIGNING` | клиент принял оффер | `OFFER_ACCEPT` |
| `APPROVED`, `COUNTER_OFFER` | `CLOSED` | клиент отказался от оффера | `OFFER_DECLINE` |
| `SIGNING` | `ISSUED` | выдача | `DISBURSE` |

`REJECTED`, `ISSUED`, `CLOSED` — конечные. Любой другой переход — `409 INVALID_TRANSITION` с текстом «Нельзя
перевести заявку из статуса «…» в «…»». Стоп-отказ этапа 1 — результат предварительного решения внутри `DRAFT`,
а не статус.

## Рамки продукта

Параметры проверяются по продукту из Catalog при создании черновика и при автосохранении. Нарушение —
`422 BUSINESS_RULE` с понятным текстом:

| Правило | Текст |
|---|---|
| сумма в `amount.min…max` | «Сумма вне рамок продукта: от 300 000 ₽ до 10 000 000 ₽» |
| срок в `term.min…max` | «Срок вне рамок продукта: от 3 до 36 мес.» |
| цель — одна из целей продукта | «Продукт «…» не подходит для этой цели» |
| продукт есть в каталоге | «Продукт не найден, выберите продукт в каталоге» |

## API

### `POST /applications`

Реализовано. Кнопка «Оформить»: создаёт черновик и ведёт в мастер.

- Компания — из пользователя. Полномочия заявителя — `shared.Authorities.Resolve` по ИНН пользователя в данных
  банка. Нет полномочий (не руководитель и нет действующей МЧД с `CREDIT_APPLICATION`) — `403 FORBIDDEN`, текст
  ошибки — причина от Company, например «Срок действия доверенности истёк 30.09.2026». Данные банка или реестр МЧД
  недоступны — `502 UPSTREAM_UNAVAILABLE`.
- Снимок полномочий (тип, номер и срок МЧД, полномочия) сохраняется в заявке. По нему потом проверяется право дать
  согласие на БКИ.
- `term` можно не передавать только у продукта с единственным сроком (овердрафт, 12 мес.), иначе —
  `400 VALIDATION_ERROR`. `goal` можно не передавать: у продукта с единственной целью она подставляется сама.
- id заявки — `a-<номер>`, нумерация с 1001. Версия черновика — 1.

Запрос:
```json
{ "productCode": "WORKING", "amount": 1500000, "term": 24, "goal": "WORKING_CAPITAL" }
```
Ответ `201`:
```json
{ "id": "a-1031", "status": "DRAFT", "version": 1, "redirect": "/applications/a-1031/edit" }
```

### `GET /applications?mine=true`

Реализовано. Блок «Мои заявки»: черновики и поданные заявки компании пользователя, недавно изменённые первыми.
Клиент всегда видит только свою компанию, `mine` принимается для совместимости. Удалённые черновики не показываются.
Пагинация — `?page=1&size=20` (размер до 100).

- `number` присваивается при отправке, у черновика — `null`.
- `actionRequired: true` — ход за клиентом: `INFO_REQUESTED`, `APPROVED`, `COUNTER_OFFER`, `SIGNING`.

```json
{
  "items": [
    { "id": "a-1024", "number": "КЗ-2026-001024", "product": { "code": "WORKING", "name": "Оборотный кредит" }, "amount": 1500000, "term": 24, "status": "INFO_REQUESTED", "statusLabel": "Нужна доп. информация", "actionRequired": true, "updatedAt": "2026-10-20T12:40:00Z" },
    { "id": "a-1031", "number": null, "product": { "code": "EXPRESS", "name": "Экспресс-кредит" }, "amount": 500000, "term": 12, "status": "DRAFT", "statusLabel": "Черновик", "actionRequired": false, "updatedAt": "2026-10-19T09:10:00Z" }
  ],
  "page": 1, "size": 20, "total": 2
}
```

### `GET /applications/{id}`

Реализовано. Полное состояние заявки: мастер по нему решает, какой этап и какие поля показать и что заблокировать.

```json
{
  "id": "a-1031",
  "number": null,
  "status": "DRAFT",
  "statusLabel": "Черновик",
  "stage": 2,
  "version": 7,
  "product": { "code": "WORKING", "name": "Оборотный кредит" },
  "params": { "amount": 1500000, "term": 24, "goal": "WORKING_CAPITAL", "purposeText": "Закупка муки и упаковки к сезону", "graceMonths": null, "overdraftLimit": null },
  "company": { "inn": "7712345678", "type": "LLC", "name": "ООО «Северная пекарня»", "ogrn": "1157746000000", "regDate": "2015-03-12", "state": "ACTIVE" },
  "authority": { "type": "MCHD", "verified": true, "validUntil": "2027-03-01" },
  "consents": { "bki": { "given": true, "at": "2026-10-20T10:05:00Z" } },
  "prelim": { "outcome": "LIMIT", "limit": 3300000, "prelimRate": 22, "amountFits": true, "choice": null, "stopReason": null },
  "prelimInvalidated": false,
  "stage2": {
    "actualAddress": "г. Москва, ул. Складская, 3",
    "turnover": { "source": "BANK_ACCOUNTS", "avgMonthlyIncome": 1840000, "periodMonths": 12 },
    "guarantor": { "personId": "p-1", "fullName": "Смирнов Олег Викторович", "bkiConsent": true },
    "truthConfirmed": false
  },
  "locks": { "paramsLocked": false, "stage2Available": true },
  "savedAt": "2026-10-20T10:12:00Z"
}
```

- `company`, `authority`, `stage2.actualAddress`, `stage2.turnover` — данные банка, только для чтения. Если данные
  банка сейчас недоступны, заявка всё равно открывается: `company`, `actualAddress`, `turnover` — `null`.
- `authority` — снимок при создании черновика: `HEAD` — руководитель, `MCHD` — по доверенности до `validUntil`.
- `stage`: 1 — параметры, согласие на БКИ и предварительное решение; 2 — открыт после лимита этапа 1, если сумма
  в лимите (`amountFits: true`) или клиент выбрал, что делать с суммой сверх лимита (`prelim/choice`). У поданной
  заявки — всегда 2.
- `prelim` — действующий результат этапа 1 (Scoring). У стоп-отказа `limit`, `prelimRate`, `amountFits` — `null`,
  заполнен `stopReason`. Результат считается устаревшим, если его посчитали для других продукта, суммы или срока:
  тогда `prelim: null` и `prelimInvalidated: true` — этап 1 нужно пройти заново.
- `stage2.guarantor` — `null`, пока поручитель не выбран; `fullName` — из кандидатов банка.
- `locks.paramsLocked` — параметры менять нельзя (заявка не черновик); `locks.stage2Available` — открыт этап 2.
- `consents.bki` — только если согласие дано.

### `PATCH /applications/{id}`

Реализовано. Автосохранение черновика: передаются только изменённые поля. Работает только в `DRAFT`.

Принимаются только `params` (`amount`, `term`, `goal`, `purposeText`, `graceMonths`, `overdraftLimit`) и
`stage2.guarantor.personId`. Любое другое поле, в том числе данные банка (`company`, `authority`,
`stage2.actualAddress`, `stage2.turnover`), — `400 VALIDATION_ERROR` с именем поля: такие данные клиент обновляет
в банке, а не в заявке. Продукт черновика не меняется: для другого продукта создаётся новый черновик.

Заголовок: `If-Match: 7`. Запрос:
```json
{
  "params": { "amount": 1200000, "purposeText": "Закупка муки" },
  "stage2": { "guarantor": { "personId": "p-1" } }
}
```
Ответ `200`:
```json
{ "version": 8, "savedAt": "2026-10-20T10:13:00Z", "prelimInvalidated": true }
```

- `null` или отсутствие поля — не менять. Снять необязательный параметр: `purposeText: ""`, `graceMonths: 0`,
  `overdraftLimit: 0`.
- Рамки продукта — как при создании. Дополнительно: отсрочка только у продукта с `graceMaxMonths > 0`, не длиннее
  его и короче срока; `overdraftLimit` — только у овердрафта и в рамках суммы продукта. Нарушение —
  `422 BUSINESS_RULE`.
- Поручитель — только у ООО и только из `guarantorCandidates` данных банка; иначе — `400 VALIDATION_ERROR` по полю
  `stage2.guarantor.personId`.
- `prelimInvalidated: true` — результат этапа 1 посчитан для других суммы, срока или продукта: этап 1 нужно пройти
  заново.
- Ошибки: `400` без `If-Match`, `412 VERSION_CONFLICT` — версия устарела, `409 INVALID_TRANSITION` — заявка
  не черновик.

### `POST /applications/{id}/consents/bki`

Реализовано. Фиксирует согласие заявителя на запрос в БКИ: кто, когда, по какой заявке. Без него этап 1
(`prelim`, зона Scoring) отвечает `403`. Только для черновика (`409` иначе).

Запрос:
```json
{ "given": true, "text": "Согласен на запрос моей кредитной истории в бюро кредитных историй" }
```
Ответ `201`:
```json
{ "given": true, "by": "u-1", "at": "2026-10-20T10:05:00Z" }
```

- Право дать согласие — по снимку полномочий: у руководителя оно есть всегда, представителю по МЧД нужно
  полномочие `BKI_CONSENT`. Нет — `403 FORBIDDEN`.
- `given: false` — `400 VALIDATION_ERROR`: отозвать согласие нельзя, без него просто нет этапа 1.
- Повторный запрос не создаёт новое согласие и возвращает уже данное (`by` и `at` первого).
- id согласия (`cs-<номер>`) передаётся в мок БКИ как `consentId`.

### `POST /applications/{id}/guarantor/consent-bki`

Реализовано. Согласие выбранного поручителя на запрос его кредитной истории. Отдельного входа у поручителя нет:
согласие фиксирует заявитель, в согласии запоминается, за какого человека оно дано.

Запрос:
```json
{ "given": true }
```
Ответ `201`:
```json
{ "given": true, "at": "2026-10-20T10:20:00Z" }
```

- Только черновик на этапе 2 (иначе `409 INVALID_TRANSITION`) с выбранным поручителем (иначе `400` по полю
  `stage2.guarantor.personId`). `given: false` — `400`.
- Если в `PATCH` выбрали другого поручителя, согласие прежнего удаляется в той же транзакции:
  `stage2.guarantor.bkiConsent` снова `false`. Повторный выбор того же человека согласие не сбрасывает.

### `DELETE /applications/{id}`

Реализовано. Удаляет черновик, только `DRAFT` (иначе `409 INVALID_TRANSITION`). Ответ `204`.

Черновик не стирается из БД: заявка переходит `DRAFT → CLOSED` с действием `DELETE_DRAFT` и отметкой удаления.
Клиенту удалённый черновик больше не виден (`404`, нет в списке), а другие зоны видят его в статусе `CLOSED`.
Очередь банка такой переход игнорирует.

## Для других зон

### `shared.Applications`

Реализовано (`internal/origination/ports.go`). Используют Scoring, Underwriting, Company (B) и Offer (A).

- `Application(id)` — заявка с текущим статусом, версией, снимком полномочий и этапом мастера. Удалённый черновик
  тоже возвращается (в статусе `CLOSED`); нет заявки — `(nil, nil)`.
- `Transition(id, to, actor, action, comment)` — переход по машине состояний с записью истории и событием
  `ApplicationStatusChanged`; недопустимый переход — `409 INVALID_TRANSITION`.
- `SetAmount(id, amount, actor)` — сумма черновика по выбору клиента на этапе 1 (`prelim/choice`), возвращает новую
  версию. Не черновик — `409`. Рамки продукта и лимит проверяет вызывающая зона (Scoring).
- `BKIConsent(appId, subject)` — согласие заявителя или поручителя на запрос в БКИ; нет — `(nil, nil)`.
