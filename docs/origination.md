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
