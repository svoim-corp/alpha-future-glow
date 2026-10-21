# Умный кредитный кабинет — спецификация API и доменные зоны

> **Частично устаревшая спецификация.** Сохранена для справки и не является эталоном поведения или подтверждением готовности API. Если в `docs/` нет других актуальных файлов с описанием нужного функционала, не используйте этот файл как эталон: сверяйте код клиента и сервера, тесты и постановку задачи. Описания модулей будут создаваться при их реализации и доработках. Ниже сохранены прежние формулировки и примеры; они могут расходиться с реализацией.

Версия: 21.10.2026. Источник структуры — холст в Design «Схемы кредитного кабинета» (схема страниц, формы заявки, схема решения, статусы). Требования — ТЗ; ссылки вида «стр. N».

Обозначения:
- **ТЗ** — требование из ТЗ, указан пункт.
- **Доп.** — наше допущение, в ТЗ нет.
- **Уточн.** — пояснение организаторов к ТЗ.
- Все внешние источники — локальные моки (стр. 10, п. 04; стр. 12, п. 5), перечислены в THIRD_PARTY_NOTICES (стр. 7).

### Пользователи — действующие клиенты банка (Уточн., 21.10.2026)

Все компании и ИП, которые входят в кабинет, уже обслуживаются в банке. Данные, нужные для кредита, у банка есть всегда, поэтому кабинет **не спрашивает их у пользователя и не просит загружать документы**:

- **Карточка компании** — реквизиты, адреса, руководитель, учредители, налоговый режим. Берётся из клиентской базы банка (АБС) и сверяется с реестрами. ИНН не вводится: компания известна по пользователю.
- **Обороты** — по счетам в банке. Выписки из 1С и других банков нет.
- **Полномочия заявителя** (руководитель или доверенность МЧД) — берутся из данных банка о представителях клиента. Номер МЧД пользователь не вводит.
- **Поручитель** (для ООО) — выбирается из руководителя и учредителей, известных банку. ФИО и ИНН не вводятся.

От клиента нужны только решения и согласия: продукт, сумма, срок, цель, согласия на запрос в БКИ, подтверждение достоверности и ответы на уточняющие вопросы банка. Признак «клиент банка» и правила продуктов «только для клиентов банка» больше не нужны: клиенты все.

---

## 1. Доменные зоны

| # | Зона | За что отвечает | Владеет данными | Основа |
|---|---|---|---|---|
| 1 | **Identity** (доступ) | Вход, сессия, роли, привязка пользователя к компании и к человеку в данных банка (`personInn`) | `User`, `Session` | стр. 5 (кто подаёт), стр. 7 (тест. учётки) |
| 2 | **Catalog** (продукты) | Линейка продуктов, их рамки, подбор по задаче, калькулятор платежа | `Product`, `Goal` | стр. 6, п. 3; стр. 13; калькулятор — Доп. |
| 3 | **Company** (клиент-юрлицо) | Профиль компании/ИП из клиентской базы банка, сверенный с реестрами: адреса, руководитель и учредители, налоговый режим, обороты по счетам в банке; полномочия заявителя (руководитель / МЧД) по данным банка | `Company`, `Person`, `Turnover` | стр. 4 (без повторного ввода), стр. 13; МЧД — Доп.; Уточн. |
| 4 | **Origination** (заявка) | Черновик, этап 1 и этап 2, проверка полноты, отправка, машина состояний, история; хранит снимок полномочий заявителя | `Application`, `Authority`, `Consent`, `StatusEvent` | стр. 6, п. 3, 5; стр. 13 |
| 5 | **Integrations** (моки) | Адаптеры к АБС банка (профиль клиента, представители, обороты), ЕГРЮЛ/ЕГРИП, «Прозрачный бизнес», БКИ, ФНС, ГИР БО, КАД/ФССП/ЕФРСБ, реестр МЧД | `CheckResult` (снимок ответа) | стр. 13; стр. 12, п. 5; часть источников — Доп. |
| 6 | **Scoring & Decision** (риск) | Предварительный лимит (этап 1), полный скоринг с факторами, кредитная логика, авторешение | `ScoreResult`, `Decision` | стр. 14–15; стр. 11, п. 06 |
| 7 | **Underwriting** (рабочее место банка) | Очередь, разбор заявки, запрос доп. информации, ручное решение, иные условия | `InfoRequest`, назначение | стр. 6, п. 4; стр. 12, п. 3; стр. 15 |
| 8 | **Offer & Contract** (оформление) | Оффер, принятие условий, имитация подписи и выдачи | `Offer`, `Contract` | стр. 11, п. 06; стр. 15 |
| 9 | **Servicing** (кредиты) | Выданные кредиты, график, платёжный календарь, напоминания в кабинете | `Loan`, `ScheduleItem` | Доп. (ТЗ заканчивает путь выдачей, стр. 15) |
| 10 | **Support** (поддержка) | Частые вопросы, обращения | `FaqItem`, `Ticket` | Доп. |
| — | **Сквозное** | In-app уведомления, аудит действий, OpenAPI | `Notification`, `AuditLog` | стр. 11, п. 05 (всё внутри продукта); стр. 7 (Swagger); стр. 9, п. 05 (enterprise) |

### Связи между зонами

```
Identity ──► Origination ◄── Catalog
                │  ▲
                ▼  │
          Integrations ──► Company
                │
                ▼
       Scoring & Decision ◄──► Underwriting
                │
                ▼
        Offer & Contract ──► Servicing
```

Правила:
- Origination — владелец машины состояний. Другие зоны меняют статус только через команды Origination (или доменные события), не напрямую в БД.
- Integrations только читает внешние моки и сохраняет снимок ответа (`CheckResult`) — решение по нему принимает Scoring & Decision.
- Scoring не знает про UI. Его результат читают и клиентский мастер (этап 1), и разбор заявки банка.
- Для MVP — модульный монолит: одна кодовая база, пакет на зону, общая БД со схемой на зону. Моки внешних сервисов — отдельный контейнер с HTTP-контрактами, чтобы интеграция была настоящей (стр. 11, п. 06; стр. 12, п. 5).

---

## 2. Машина состояний заявки

| Код | Статус | Кто переводит дальше |
|---|---|---|
| `DRAFT` | Черновик (в т.ч. этап 1 пройден) | клиент |
| `SUBMITTED` | Подана | система |
| `DATA_CHECK` | Проверка данных | система / сотрудник |
| `INFO_REQUESTED` | Нужна доп. информация | клиент → `DATA_CHECK` |
| `SCORING` | Обогащение и скоринг | система |
| `ON_DECISION` | На решении | система (авто) / сотрудник |
| `APPROVED` / `COUNTER_OFFER` / `REJECTED` | Одобрено / Иные условия / Отказ | клиент принимает оффер |
| `SIGNING` | Подписание | клиент |
| `ISSUED` | Выдано | — |
| `CLOSED` | Закрыта (клиент отказался от оффера или удалил черновик) | — |

Стоп-отказ этапа 1 — это результат предварительного решения внутри `DRAFT` (`prelim.outcome = STOP`), а не статус. Набор статусов — Доп.; ТЗ задаёт этапы и исходы (стр. 13–15). Диаграмма состояний обязательна к сдаче (стр. 7).

Недопустимый переход → `409 Conflict` с кодом `INVALID_TRANSITION`.

---

## 3. Соглашения API

- Префикс: `/api/v1`. Формат: `application/json; charset=utf-8`. Загрузки файлов нет: данные клиента банк берёт сам (Уточн.).
- Аутентификация: httpOnly cookie `session` после `/auth/login`. Роли: `CLIENT`, `BANK_OFFICER`. Клиент видит только заявки своей компании; `/bank/*` — только `BANK_OFFICER`.
- Деньги — целые рубли (`number`), ставки — проценты годовых (`number`, `22.5`), сроки — месяцы. Даты — ISO 8601 (`2026-10-20`, `2026-10-20T15:57:00Z`).
- Оптимистичная блокировка: заявка возвращает `version`, изменения передают её в заголовке `If-Match`. Несовпадение → `412`.
- Идемпотентность: команды с побочным эффектом (`submit`, `sign`, `disburse`, `decision`) принимают заголовок `Idempotency-Key`.
- Пагинация: `?page=1&size=20` → `{ "items": [], "page": 1, "size": 20, "total": 134 }`.

### Формат ошибки

```json
{
  "code": "VALIDATION_ERROR",
  "message": "Не все поля заполнены",
  "fields": [
    { "field": "stage2.guarantor.personId", "code": "REQUIRED", "message": "Выберите поручителя" }
  ],
  "traceId": "b7f3c2a1"
}
```

| HTTP | `code` | Когда |
|---|---|---|
| 400 | `VALIDATION_ERROR` | Неверные или пропущенные поля |
| 401 | `UNAUTHORIZED` | Нет сессии |
| 403 | `FORBIDDEN` | Чужая заявка, не та роль, нет полномочий (по данным банка заявитель не руководитель и действующей МЧД нет) |
| 404 | `NOT_FOUND` | Нет объекта |
| 409 | `INVALID_TRANSITION` | Действие недопустимо в текущем статусе |
| 412 | `VERSION_CONFLICT` | Заявку изменили параллельно |
| 422 | `BUSINESS_RULE` | Нарушено правило продукта (сумма или срок вне рамок, отсрочка длиннее разрешённой) |
| 502 | `UPSTREAM_UNAVAILABLE` | Мок внешнего сервиса не ответил |

### Общие объекты

**`ApplicationShort`** — строка в списках:

```json
{
  "id": "a-1024",
  "number": "КЗ-2026-001024",
  "product": { "code": "WORKING", "name": "Оборотный кредит" },
  "amount": 1500000,
  "term": 24,
  "status": "INFO_REQUESTED",
  "statusLabel": "Нужна доп. информация",
  "actionRequired": true,
  "updatedAt": "2026-10-20T12:40:00Z"
}
```

**`Money range`** — диапазон для платежа и ставки: `{ "min": 76300, "max": 82300 }`.

---

## 4. Эндпоинты по страницам

### 4.1 Вход — `/login`
Зона: Identity. Регистрации нет (ответ организаторов). Тестовые учётки — в seed и README (стр. 7).

#### `POST /auth/login`
Проверяет логин и пароль, создаёт сессию, говорит фронту, куда перейти.

Запрос:
```json
{ "login": "smirnov", "password": "demo123" }
```
Ответ `200` (+ `Set-Cookie: session=...`):
```json
{
  "user": {
    "id": "u-1",
    "fullName": "Смирнов Олег Викторович",
    "position": "Генеральный директор",
    "role": "CLIENT",
    "companyId": "c-1"
  },
  "redirect": "/"
}
```
Для сотрудника `role: "BANK_OFFICER"`, `companyId: null`, `redirect: "/bank/queue"`. Ошибка: `401 UNAUTHORIZED`.

#### `POST /auth/logout`
Завершает сессию. Тела нет. Ответ `204`.

#### `GET /auth/me`
Текущий пользователь для шапки и проверки доступа на фронте.

Ответ `200`:
```json
{
  "id": "u-2",
  "fullName": "Петрова Анна Сергеевна",
  "position": "Главный бухгалтер",
  "role": "CLIENT",
  "company": { "id": "c-1", "inn": "7712345678", "name": "ООО «Северная пекарня»" }
}
```
Признака `isBankClient` нет: все пользователи — клиенты банка (Уточн.).

### 4.2 Шапка клиента (на всех страницах)

#### `GET /notifications?unread=true`
Уведомления внутри кабинета: запрос банка, решение, напоминание о платеже. Замена SMS и почте (стр. 11, п. 05).

Ответ `200`:
```json
{
  "unreadCount": 2,
  "items": [
    {
      "id": "n-15",
      "type": "INFO_REQUESTED",
      "title": "Банк просит дополнить заявку",
      "text": "Уточните фактический адрес склада",
      "link": "/applications/a-1024",
      "read": false,
      "createdAt": "2026-10-20T12:40:00Z"
    }
  ]
}
```
`type`: `INFO_REQUESTED`, `DECISION`, `PAYMENT_DUE`, `TICKET_REPLY`.

#### `POST /notifications/{id}/read`
Отмечает уведомление прочитанным. Тела нет. Ответ `204`.

### 4.3 Главная — `/`
Зоны: Origination, Servicing. ТЗ: стр. 4 (прозрачность статуса), стр. 6, п. 5.

#### `GET /applications?mine=true`
Блок «Мои заявки»: черновики и поданные заявки компании, новые сверху.

Ответ `200`:
```json
{
  "items": [
    { "id": "a-1024", "number": "КЗ-2026-001024", "product": { "code": "WORKING", "name": "Оборотный кредит" }, "amount": 1500000, "term": 24, "status": "INFO_REQUESTED", "statusLabel": "Нужна доп. информация", "actionRequired": true, "updatedAt": "2026-10-20T12:40:00Z" },
    { "id": "a-1031", "number": null, "product": { "code": "EXPRESS", "name": "Экспресс-кредит" }, "amount": 500000, "term": 12, "status": "DRAFT", "statusLabel": "Черновик", "actionRequired": false, "updatedAt": "2026-10-19T09:10:00Z" }
  ],
  "page": 1, "size": 20, "total": 2
}
```
`number` присваивается при отправке, у черновика — `null`.

#### `GET /loans/summary`
Есть ли выданные кредиты и ближайший платёж — для перехода в «Мои кредиты».

Ответ `200`:
```json
{ "activeLoans": 1, "nextPayment": { "date": "2026-11-15", "amount": 77800 } }
```

### 4.4 Каталог продуктов — `/products`
Зона: Catalog. ТЗ: стр. 6, п. 3; стр. 13. Линейка — мок (Доп.).

#### `GET /products`
Список карточек каталога.

Ответ `200`:
```json
[
  {
    "code": "WORKING",
    "name": "Оборотный кредит",
    "shortText": "Закупки, зарплата, налоги, аренда",
    "goals": ["WORKING_CAPITAL"],
    "amount": { "min": 300000, "max": 10000000 },
    "term": { "min": 3, "max": 36 },
    "rate": { "min": 20, "max": 28 }
  }
]
```

#### `GET /products/{code}`
Все условия для боковой панели «Условия».

Ответ `200`:
```json
{
  "code": "INVEST",
  "name": "Инвестиционный кредит",
  "description": "Оборудование, помещение, расширение бизнеса",
  "goals": ["EQUIPMENT", "REAL_ESTATE", "EXPANSION"],
  "amount": { "min": 1000000, "max": 30000000 },
  "term": { "min": 12, "max": 84 },
  "rate": { "min": 18, "max": 26 },
  "graceMaxMonths": 6,
  "requirements": {
    "minBusinessAgeMonths": 12,
    "guarantorForLlc": true
  },
  "limitRule": "До 1,5 среднемесячных поступлений на счета в банке",
  "whatYouNeed": ["Согласие на запрос в БКИ", "Для ООО — поручитель из руководителя или учредителей и его согласие на запрос в БКИ"]
}
```
`whatYouNeed` перечисляет только действия клиента. Данные компании и обороты банк берёт сам (Уточн.).

#### `GET /products/goals`
Цели для блока «Подбор по задаче».

Ответ `200`:
```json
[
  { "code": "WORKING_CAPITAL", "name": "Пополнить оборотные средства" },
  { "code": "EQUIPMENT", "name": "Купить оборудование" },
  { "code": "CASH_GAP", "name": "Закрыть кассовый разрыв" }
]
```

#### `POST /products/match`
Подбор продукта по задаче. Подсвечивает подходящие карточки.

Запрос:
```json
{ "goal": "WORKING_CAPITAL", "amount": 1500000, "term": 24 }
```
Ответ `200`:
```json
[
  { "code": "WORKING", "fits": true, "reason": "Подходит по цели, сумме и сроку" },
  { "code": "OVERDRAFT", "fits": false, "reason": "Подходит для кассовых разрывов, а не для оборотных средств" },
  { "code": "EXPRESS", "fits": false, "reason": "Сумма больше 1 000 000 ₽" }
]
```

#### `POST /applications`
Кнопка «Оформить»: создаёт черновик с параметрами из подбора и ведёт в мастер. Компания берётся из пользователя. Её карточка и полномочия заявителя подтягиваются из данных банка при создании черновика (Уточн.).

Запрос:
```json
{ "productCode": "WORKING", "amount": 1500000, "term": 24, "goal": "WORKING_CAPITAL" }
```
Ответ `201`:
```json
{ "id": "a-1031", "status": "DRAFT", "version": 1, "redirect": "/applications/a-1031/edit" }
```
Ошибки:
- `422 BUSINESS_RULE` — сумма или срок вне рамок продукта.
- `403 FORBIDDEN` — по данным банка у заявителя нет полномочий подавать заявку: он не руководитель и действующей МЧД с полномочием `CREDIT_APPLICATION` нет. Текст ошибки объясняет, что сделать: оформить МЧД или подать заявку от имени руководителя.
- `502 UPSTREAM_UNAVAILABLE` — данные банка или реестра МЧД недоступны.

### 4.5 Мастер заявки — `/applications/:id/edit`
Зоны: Origination (владелец), Identity, Company, Integrations, Scoring, Catalog. ТЗ: стр. 13 (сумма, срок, цель, сведения о клиенте; проверка полноты), стр. 4 (без повторного ввода), стр. 6, п. 5 (сохранение данных). Двухэтапность, МЧД, поручитель — Доп.

Сведения о клиенте (карточка компании, адреса, обороты, полномочия, кандидаты в поручители) мастер только показывает. Клиент их не вводит и не правит (Уточн.).

#### `GET /applications/{id}`
Полное состояние заявки. Мастер по нему решает, какой этап и какие поля показать и что заблокировать.

Ответ `200`:
```json
{
  "id": "a-1031",
  "number": null,
  "status": "DRAFT",
  "stage": 2,
  "version": 7,
  "product": { "code": "WORKING", "name": "Оборотный кредит" },
  "params": {
    "amount": 1500000,
    "term": 24,
    "goal": "WORKING_CAPITAL",
    "purposeText": "Закупка муки и упаковки к сезону",
    "graceMonths": null,
    "overdraftLimit": null
  },
  "company": {
    "inn": "7712345678",
    "type": "LLC",
    "name": "ООО «Северная пекарня»",
    "ogrn": "1157746000000",
    "regDate": "2015-03-12",
    "state": "ACTIVE"
  },
  "authority": { "type": "MCHD", "verified": true, "validUntil": "2027-03-01" },
  "consents": { "bki": { "given": true, "at": "2026-10-20T10:05:00Z" } },
  "prelim": {
    "outcome": "LIMIT",
    "limit": 3000000,
    "prelimRate": 22,
    "amountFits": true,
    "choice": null
  },
  "stage2": {
    "actualAddress": "г. Москва, ул. Складская, 3",
    "turnover": { "source": "BANK_ACCOUNTS", "avgMonthlyIncome": 1840000, "periodMonths": 12 },
    "guarantor": { "personId": "p-1", "fullName": "Смирнов Олег Викторович", "bkiConsent": true },
    "truthConfirmed": false
  },
  "locks": { "stage2Available": true },
  "savedAt": "2026-10-20T10:12:00Z"
}
```
- `company`, `authority`, `stage2.actualAddress` и `stage2.turnover` — данные банка, только для чтения (Уточн.).
- `authority.type`: `HEAD` — заявитель руководитель, `MCHD` — действует по доверенности, известной банку.
- `stage2.guarantor` — `null`, пока поручитель не выбран. Для ИП поля нет.

#### `PATCH /applications/{id}`
Автосохранение черновика. Передаются только изменённые поля. Работает только в `DRAFT`.

Заголовок: `If-Match: 7`. Запрос:
```json
{
  "params": { "amount": 1200000, "purposeText": "Закупка муки" }
}
```
Ответ `200`:
```json
{ "version": 8, "savedAt": "2026-10-20T10:13:00Z", "prelimInvalidated": true }
```
`prelimInvalidated: true` — изменилась сумма, срок или продукт, этап 1 нужно пройти заново.

Принимаются только `params` (`amount`, `term`, `goal`, `purposeText`, `graceMonths`, `overdraftLimit`) и `stage2.guarantor.personId`. Данные банка (`company`, `authority`, `stage2.actualAddress`, `stage2.turnover`) через PATCH не меняются: попытка — `400 VALIDATION_ERROR`.

Ошибки: `412 VERSION_CONFLICT`, `409 INVALID_TRANSITION` (не черновик).

#### Этап 1 — предварительное решение

Компания и полномочия заявителя определяются при создании черновика (`POST /applications`) по данным банка. Отдельных шагов «ввести ИНН» и «ввести номер МЧД» нет (Уточн.). Этап 1 — это параметры кредита, согласие на БКИ и предварительное решение.

##### `POST /calculator/payment`
Калькулятор под суммой и сроком. Ставку до проверки не знаем, поэтому платёж — диапазоном по ставке продукта.

Запрос:
```json
{ "productCode": "INVEST", "amount": 6000000, "term": 60, "graceMonths": 6 }
```
Ответ `200`:
```json
{
  "rate": { "min": 18, "max": 26 },
  "payment": { "min": 162900, "max": 189600 },
  "gracePayment": { "min": 90000, "max": 130000, "months": 6 },
  "withinProductLimits": true
}
```
`gracePayment` — только при отсрочке (проценты без основного долга). Для обычных продуктов — `null`.

##### `POST /applications/{id}/consents/bki`
Фиксирует согласие на запрос в БКИ: кто, когда, по какой заявке.

Запрос:
```json
{ "given": true, "text": "Согласен на запрос моей кредитной истории в бюро кредитных историй" }
```
Ответ `201`:
```json
{ "given": true, "by": "u-1", "at": "2026-10-20T10:05:00Z" }
```
Ошибка: `403 FORBIDDEN`, если по доверенности у заявителя нет полномочия `BKI_CONSENT`.

##### `POST /applications/{id}/prelim`
Запускает этап 1: ЕГРЮЛ/ЕГРИП, БКИ, ФНС, КАД/ФССП/ЕФРСБ, ГИР БО (только ООО), обороты по счетам в банке. Возвращает лимит или стоп-отказ. Работает синхронно: моки отвечают быстро.

Запрос: тела нет (данные берутся из черновика и из банка). Требует согласие на БКИ.

Ответ `200` — лимит:
```json
{
  "outcome": "LIMIT",
  "requestedAmount": 1500000,
  "limit": 3000000,
  "prelimRate": 22,
  "prelimPayment": 77800,
  "amountFits": true,
  "checks": [
    { "source": "EGRUL", "status": "OK", "label": "Компания действует, 11 лет" },
    { "source": "BANK_ACCOUNTS", "status": "OK", "label": "Обороты по счетам учтены" },
    { "source": "BKI", "status": "OK", "label": "Просрочек нет" },
    { "source": "FNS", "status": "OK", "label": "Налоговой задолженности нет" },
    { "source": "COURTS", "status": "WARN", "label": "1 арбитражное дело, некритично" },
    { "source": "GIRBO", "status": "OK", "label": "Отчётность учтена" }
  ]
}
```
- `limit` считается по оборотам на счетах в банке, без выписок (Уточн.).
- `amountFits: true` — сумма в пределах лимита, этап 2 открывается сразу.
- `amountFits: false` — сумма больше лимита, клиент выбирает дальше (`prelim/choice`).
- Цифры ГИР БО клиенту не показываем, только «учтена».

Ответ `200` — стоп-отказ:
```json
{
  "outcome": "STOP",
  "stopReason": "BUSINESS_TOO_YOUNG",
  "stopText": "Бизнес зарегистрирован меньше 6 месяцев назад",
  "whatToDo": ["Вернуться, когда бизнесу будет 6 месяцев", "Посмотреть экспресс-кредит после этого срока"]
}
```
`stopReason`: `COMPANY_INACTIVE`, `BUSINESS_TOO_YOUNG`, `BKI_OVERDUE`, `CRITICAL_RISK`. Ошибки: `403` (нет согласия или полномочий), `502`.

##### `POST /applications/{id}/prelim/choice`
Выбор клиента, если сумма больше лимита (`amountFits: false`): уменьшить сумму до лимита или оставить запрошенную. Во втором случае решение примет сотрудник банка, вероятны иные условия. Доп.

Запрос:
```json
{ "choice": "REDUCE_AMOUNT", "amount": 3000000 }
```
или
```json
{ "choice": "KEEP_AMOUNT" }
```
Ответ `200`:
```json
{ "stage": 2, "params": { "amount": 3000000, "term": 24 }, "version": 9 }
```
Ошибки: `409 INVALID_TRANSITION`, если `amountFits: true` или этап 1 не пройден; `422 BUSINESS_RULE`, если `amount` больше лимита или вне рамок продукта.

#### Этап 2 — полная заявка

Этап 2 — это проверка данных банка, выбор поручителя, согласие поручителя и подтверждение достоверности. Документы не загружаются (Уточн.).

##### `GET /applications/{id}/company-profile`
Данные о компании из клиентской базы банка, сверенные с реестрами. Только для чтения. Если данные устарели, клиент обновляет их в банке, а не в заявке.

Ответ `200`:
```json
{
  "source": ["BANK", "EGRUL", "TRANSPARENT_BUSINESS"],
  "regAddress": "г. Москва, ул. Пекарская, 7",
  "actualAddress": "г. Москва, ул. Складская, 3",
  "okved": { "code": "10.71", "name": "Производство хлеба" },
  "head": { "personId": "p-1", "fullName": "Смирнов Олег Викторович" },
  "founders": [{ "personId": "p-1", "fullName": "Смирнов Олег Викторович", "share": 100 }],
  "guarantorCandidates": [{ "personId": "p-1", "fullName": "Смирнов Олег Викторович", "roles": ["HEAD", "FOUNDER"] }],
  "taxRegime": "USN_INCOME_MINUS_EXPENSES",
  "turnover": { "source": "BANK_ACCOUNTS", "avgMonthlyIncome": 1840000, "periodMonths": 12 },
  "girbo": { "accounted": true }
}
```
`guarantorCandidates` — руководитель и учредители-физлица. Для ИП — пустой массив.

##### `PATCH /applications/{id}` (этап 2)
Тот же эндпоинт автосохранения: выбор поручителя из `guarantorCandidates`.

Запрос:
```json
{
  "stage2": {
    "guarantor": { "personId": "p-1" }
  }
}
```
Поручитель обязателен для ООО, для ИП поле не показывается. ФИО и ИНН поручителя банк знает, клиент их не вводит (Уточн.). `personId` не из списка кандидатов — `400 VALIDATION_ERROR`.

##### `POST /applications/{id}/guarantor/consent-bki`
Согласие поручителя на запрос его кредитной истории.

Запрос:
```json
{ "given": true }
```
Ответ `201`:
```json
{ "given": true, "at": "2026-10-20T10:20:00Z" }
```

##### `POST /applications/{id}/validate`
Проверка полноты до отправки. Мастер подсвечивает незаполненное.

Ответ `200`:
```json
{
  "valid": false,
  "missing": [
    { "field": "stage2.guarantor.personId", "message": "Выберите поручителя" },
    { "field": "stage2.guarantor.bkiConsent", "message": "Нужно согласие поручителя на запрос в БКИ" },
    { "field": "stage2.truthConfirmed", "message": "Подтвердите достоверность сведений" }
  ]
}
```

##### `POST /applications/{id}/submit`
Отправка заявки. Переводит в `SUBMITTED`, присваивает номер и запускает банковский конвейер.

Заголовки: `If-Match`, `Idempotency-Key`. Запрос:
```json
{ "truthConfirmed": true }
```
Ответ `200`:
```json
{ "id": "a-1031", "number": "КЗ-2026-001031", "status": "SUBMITTED", "redirect": "/applications/a-1031" }
```
Ошибки: `400 VALIDATION_ERROR` (как в `validate`), `409 INVALID_TRANSITION`.

##### `DELETE /applications/{id}`
Удаляет черновик. Только `DRAFT`. Ответ `204`.

### 4.6 Карточка заявки — `/applications/:id`
Зоны: Origination, Underwriting, Offer. ТЗ: стр. 4–5 (прозрачный статус), стр. 13, 15 (уточнение информации, исходы), стр. 11, п. 05.

Сводка данных — тот же `GET /applications/{id}`.

#### `GET /applications/{id}/timeline`
Таймлайн и понятный клиенту текст «что сейчас происходит».

Ответ `200`:
```json
{
  "current": {
    "status": "INFO_REQUESTED",
    "label": "Нужна доп. информация",
    "whatsHappening": "Банк проверил заявку и просит уточнить данные",
    "nextStep": "Ответьте на запрос банка ниже",
    "actor": "CLIENT"
  },
  "steps": [
    { "status": "SUBMITTED", "label": "Подана", "at": "2026-10-20T10:25:00Z", "state": "DONE" },
    { "status": "DATA_CHECK", "label": "Проверка данных", "at": "2026-10-20T10:26:00Z", "state": "DONE" },
    { "status": "INFO_REQUESTED", "label": "Нужна доп. информация", "at": "2026-10-20T12:40:00Z", "state": "CURRENT" },
    { "status": "SCORING", "label": "Обогащение и скоринг", "at": null, "state": "PENDING" },
    { "status": "ON_DECISION", "label": "На решении", "at": null, "state": "PENDING" }
  ]
}
```
`actor` — чей ход: `CLIENT` или `BANK`.

#### `GET /applications/{id}/info-requests`
Открытые и закрытые запросы банка. Банк запрашивает только пояснения, которых нет в его данных. Реквизиты, адреса, обороты и документы, которые у банка уже есть, не запрашиваются (Уточн.).

Ответ `200`:
```json
[
  {
    "id": "r-7",
    "status": "OPEN",
    "comment": "Обороты по счетам расходятся с отчётностью",
    "items": [
      { "key": "turnover_note", "type": "TEXT", "label": "Почему выручка за 2025 год ниже оборотов по счёту" },
      { "key": "purpose_details", "type": "TEXT", "label": "У каких поставщиков планируете закупку" }
    ],
    "createdAt": "2026-10-20T12:40:00Z"
  }
]
```

#### `POST /applications/{id}/info-requests/{reqId}/answer`
Ответ клиента на запрос. Заявка возвращается в `DATA_CHECK`.

Запрос:
```json
{
  "answers": [
    { "key": "turnover_note", "value": "Часть поступлений — возврат займа учредителя" },
    { "key": "purpose_details", "value": "Мелькомбинат «Заря», упаковка — «ПакСервис»" }
  ]
}
```
Ответ `200`:
```json
{ "requestStatus": "ANSWERED", "applicationStatus": "DATA_CHECK" }
```

#### `GET /applications/{id}/offer`
Решение и оффер. Появляется после решения.

Ответ `200` — одобрено или иные условия:
```json
{
  "decision": "COUNTER_OFFER",
  "requested": { "productCode": "WORKING", "amount": 1500000, "term": 24 },
  "offered": { "productCode": "WORKING", "amount": 1200000, "term": 24, "rate": 21.5, "monthlyPayment": 62200 },
  "explanation": "Сумма снижена по уровню долговой нагрузки",
  "validUntil": "2026-10-27"
}
```
Отказ:
```json
{ "decision": "REJECTED", "reason": "Высокая долговая нагрузка по данным БКИ", "alternatives": [{ "productCode": "EXPRESS", "maxAmount": 500000 }] }
```
До решения: `404 NOT_FOUND`.

#### `POST /applications/{id}/offer/accept`
Принять условия → `SIGNING`. Тела нет.

Ответ `200`:
```json
{ "status": "SIGNING", "redirect": "/applications/a-1031/sign" }
```

#### `POST /applications/{id}/offer/decline`
Отказаться от оффера → `CLOSED`.

Запрос:
```json
{ "reason": "Нашли финансирование в другом месте" }
```
Ответ `200`:
```json
{ "status": "CLOSED" }
```

### 4.7 Подписание и выдача — `/applications/:id/sign`
Зона: Offer & Contract. ТЗ: стр. 11, п. 06; стр. 15.

#### `GET /applications/{id}/contract`
Итоговые условия и документы к подписанию.

Ответ `200`:
```json
{
  "terms": { "productCode": "WORKING", "amount": 1200000, "term": 24, "rate": 21.5, "monthlyPayment": 62200, "firstPaymentDate": "2026-11-20" },
  "documents": [
    { "id": "d-1", "title": "Кредитный договор", "url": "/api/v1/applications/a-1031/contract/documents/d-1" },
    { "id": "d-2", "title": "Договор поручительства", "url": "/api/v1/applications/a-1031/contract/documents/d-2" }
  ],
  "signed": false,
  "disbursed": false
}
```
Договор поручительства — только для ООО.

#### `GET /applications/{id}/contract/documents/{docId}`
Документ к подписанию, сформированный банком по условиям оффера. Это не библиотека файлов клиента. Ответ `200` — печатная форма `text/html; charset=utf-8` с `Content-Disposition: inline`, фронт открывает её в новой вкладке. Формат печатной формы — Доп. Ошибка: `404`, если документа нет или оффер ещё не принят.

#### `POST /applications/{id}/contract/sign`
Имитация подписи без SMS (модалка «Подтверждение подписи»).

Заголовок: `Idempotency-Key`. Запрос:
```json
{ "confirm": true }
```
Ответ `200`:
```json
{ "signed": true, "signedAt": "2026-10-20T14:00:00Z", "signatureId": "SIM-8F2A" }
```

#### `POST /applications/{id}/contract/disburse`
Имитация выдачи → `ISSUED`. Создаёт кредит и график платежей.

Заголовок: `Idempotency-Key`. Тела нет. Ответ `200`:
```json
{ "status": "ISSUED", "loanId": "l-55", "disbursedAt": "2026-10-20T14:01:00Z", "accountMasked": "•••• 4471" }
```
Ошибка: `409`, если договор не подписан.

### 4.8 Мои кредиты — `/credits`
Зона: Servicing. Доп.

#### `GET /loans`
Список выданных кредитов.

Ответ `200`:
```json
[
  {
    "id": "l-55",
    "product": { "code": "WORKING", "name": "Оборотный кредит" },
    "amount": 1200000,
    "balance": 1200000,
    "rate": 21.5,
    "term": 24,
    "issuedAt": "2026-10-20",
    "nextPayment": { "date": "2026-11-20", "amount": 62200 },
    "status": "ACTIVE"
  }
]
```

#### `GET /loans/{id}`
Детали кредита: условия договора и история платежей.

Ответ `200`:
```json
{
  "id": "l-55",
  "applicationId": "a-1031",
  "contractNumber": "КД-2026-0055",
  "terms": { "amount": 1200000, "term": 24, "rate": 21.5, "monthlyPayment": 62200 },
  "payments": [
    { "date": "2026-11-20", "principal": 40700, "interest": 21500, "total": 62200, "status": "PLANNED" }
  ]
}
```

### 4.9 Платёжный календарь — `/credits/calendar`
Зона: Servicing. Доп. Напоминания — через `/notifications`.

#### `GET /loans/schedule?from=2026-11-01&to=2026-12-31`
Платежи всех кредитов за период для календаря.

Ответ `200`:
```json
[
  { "date": "2026-11-20", "loanId": "l-55", "productName": "Оборотный кредит", "amount": 62200, "status": "PLANNED" },
  { "date": "2026-12-20", "loanId": "l-55", "productName": "Оборотный кредит", "amount": 62200, "status": "PLANNED" }
]
```

#### `GET /loans/schedule/upcoming?limit=3`
Ближайшие платежи. Формат элементов тот же.

### 4.10 Поддержка — `/support`
Зона: Support. Доп.

#### `GET /support/faq?step=APPLICATION`
Частые вопросы по шагу пути. `step`: `PRODUCT`, `APPLICATION`, `DOCUMENTS`, `STATUS`, `DECISION`.

Ответ `200`:
```json
[
  { "id": "f-3", "step": "APPLICATION", "question": "Нужно ли загружать документы?", "answer": "Нет. Данные о компании и оборотах банк берёт из ваших счетов и реестров" }
]
```

#### `POST /support/tickets`
Новое обращение. Со страницы заявки `applicationId` подставляется сам.

Запрос:
```json
{ "subject": "Устаревший фактический адрес", "text": "В заявке старый адрес склада, мы переехали", "applicationId": "a-1031" }
```
Ответ `201`:
```json
{ "id": "t-12", "status": "OPEN", "createdAt": "2026-10-20T11:00:00Z" }
```

#### `GET /support/tickets`
Мои обращения.

Ответ `200`:
```json
[
  { "id": "t-12", "subject": "Устаревший фактический адрес", "status": "ANSWERED", "reply": "Обновили адрес в данных банка, в заявке он уже новый", "applicationId": "a-1031", "updatedAt": "2026-10-20T11:30:00Z" }
]
```

### 4.11 Очередь заявок — `/bank/queue`
Зона: Underwriting. Роль `BANK_OFFICER`. ТЗ: стр. 6, п. 4; стр. 12, п. 3.

#### `GET /bank/applications/stats`
Счётчики над таблицей.

Ответ `200`:
```json
{ "SUBMITTED": 3, "DATA_CHECK": 2, "INFO_REQUESTED": 4, "SCORING": 1, "ON_DECISION": 5, "decidedToday": 7 }
```

#### `GET /bank/applications?status=ON_DECISION&product=WORKING&risk=MEDIUM&decisionMode=MANUAL&page=1&size=20&sort=-submittedAt`
Таблица заявок с фильтрами.

Ответ `200`:
```json
{
  "items": [
    {
      "id": "a-1024",
      "number": "КЗ-2026-001024",
      "company": { "inn": "7712345678", "name": "ООО «Северная пекарня»" },
      "product": { "code": "WORKING", "name": "Оборотный кредит" },
      "amount": 1500000,
      "term": 24,
      "status": "ON_DECISION",
      "score": 640,
      "riskBand": "MEDIUM",
      "decisionMode": "MANUAL",
      "assignee": null,
      "submittedAt": "2026-10-20T10:25:00Z"
    }
  ],
  "page": 1, "size": 20, "total": 5
}
```
`decisionMode`: `AUTO` — решено автоматически, `MANUAL` — ждёт сотрудника.

#### `POST /bank/applications/{id}/assign`
Взять заявку в работу. Тела нет.

Ответ `200`:
```json
{ "assignee": { "id": "u-9", "fullName": "Иванова Е. П." } }
```
Ошибка: `409`, если уже взял другой сотрудник.

### 4.12 Разбор заявки — `/bank/applications/:id`
Зоны: Underwriting, Integrations, Scoring & Decision. ТЗ: стр. 13 (полнота, обогащение), стр. 14 (скоринг с факторами), стр. 15 (иные условия, запрос информации), стр. 6, п. 4.

#### `GET /bank/applications/{id}`
Анкета глазами сотрудника. Отличие от клиентской версии — цифры ГИР БО, предварительный результат и заявитель.

Ответ `200`:
```json
{
  "id": "a-1024",
  "number": "КЗ-2026-001024",
  "status": "ON_DECISION",
  "applicant": { "fullName": "Петрова Анна Сергеевна", "authority": "MCHD", "mchdNumber": "a1b2c3d4-5e6f-4a7b-8c9d-0e1f2a3b4c5d", "mchdValidUntil": "2027-03-01" },
  "company": { "inn": "7712345678", "name": "ООО «Северная пекарня»", "type": "LLC", "regDate": "2015-03-12", "okved": "10.71", "taxRegime": "USN_INCOME_MINUS_EXPENSES", "actualAddress": "г. Москва, ул. Складская, 3", "clientSince": "2016-02-01" },
  "request": { "productCode": "WORKING", "amount": 1500000, "term": 24, "goal": "WORKING_CAPITAL", "purposeText": "Закупка муки и упаковки к сезону" },
  "turnover": { "source": "BANK_ACCOUNTS", "avgMonthlyIncome": 1840000, "periodMonths": 12 },
  "girbo": { "year": 2025, "revenue": 21400000, "netProfit": 2100000 },
  "guarantor": { "personId": "p-1", "fullName": "Смирнов Олег Викторович", "inn": "771234567890", "bkiConsent": true },
  "prelim": { "limit": 3000000, "prelimRate": 22, "amountFits": true, "choice": null }
}
```

#### `GET /bank/applications/{id}/completeness`
Проверка полноты и корректности.

Ответ `200`:
```json
{
  "complete": true,
  "issues": [
    { "severity": "WARN", "code": "TURNOVER_MISMATCH", "text": "Обороты по счетам в банке выше выручки в ГИР БО на 7%" }
  ]
}
```
`severity`: `INFO`, `WARN`, `BLOCKER`.

#### `GET /bank/applications/{id}/checks`
Результаты внешних проверок с временем запроса.

Ответ `200`:
```json
[
  { "source": "BANK", "status": "OK", "requestedAt": "2026-10-20T10:26:00Z", "summary": "Клиент банка с 2016 года, обороты 1,84 млн ₽/мес.", "details": { "clientSince": "2016-02-01", "avgMonthlyIncome": 1840000 } },
  { "source": "EGRUL", "status": "OK", "requestedAt": "2026-10-20T10:26:00Z", "summary": "Действует, 11 лет", "details": { "state": "ACTIVE", "regDate": "2015-03-12" } },
  { "source": "BKI", "status": "OK", "requestedAt": "2026-10-20T10:26:00Z", "summary": "2 кредита, просрочек нет", "details": { "activeLoans": 2, "totalDebt": 900000, "overdueDays": 0 } },
  { "source": "TRANSPARENT_BUSINESS", "status": "OK", "requestedAt": "2026-10-20T10:26:00Z", "summary": "Нет признаков массового адреса" },
  { "source": "COURTS", "status": "WARN", "requestedAt": "2026-10-20T10:26:00Z", "summary": "1 арбитражное дело, 120 000 ₽" }
]
```
`status`: `OK`, `WARN`, `FAIL`, `UNAVAILABLE`.

#### `POST /bank/applications/{id}/checks/refresh`
Перезапросить проверки. Запрос (необязательно):
```json
{ "sources": ["BKI", "COURTS"] }
```
Ответ `200` — тот же массив, что в `GET .../checks`.

#### `GET /bank/applications/{id}/score`
Итоговый балл и вклад каждого фактора (стр. 14: каждому признаку — своё влияние на оценку).

Ответ `200`:
```json
{
  "score": 640,
  "riskBand": "MEDIUM",
  "autoDecisionEligible": false,
  "autoDecisionReason": "Пограничный балл",
  "factors": [
    { "code": "BUSINESS_AGE", "label": "Возраст бизнеса 11 лет", "impact": 60 },
    { "code": "BKI_DISCIPLINE", "label": "Без просрочек", "impact": 80 },
    { "code": "DEBT_LOAD", "label": "Долговая нагрузка 47% от оборота", "impact": -70 },
    { "code": "TURNOVER_STABILITY", "label": "Обороты по счетам в банке стабильны 12 мес.", "impact": 40 },
    { "code": "COURTS", "label": "Арбитражное дело", "impact": -20 }
  ],
  "suggested": { "decision": "COUNTER_OFFER", "amount": 1200000, "rate": 21.5 },
  "modelVersion": "rules-v1"
}
```

#### `POST /bank/applications/{id}/info-requests`
Запрос информации у клиента → `INFO_REQUESTED` и уведомление клиенту.

Запрос:
```json
{
  "comment": "Обороты по счетам расходятся с отчётностью",
  "items": [
    { "key": "turnover_note", "type": "TEXT", "label": "Почему выручка за 2025 год ниже оборотов по счёту" }
  ]
}
```
Пункты запроса — только пояснения. Данные, которые есть у банка, и загрузка документов не запрашиваются (Уточн.).
Ответ `201`:
```json
{ "id": "r-7", "status": "OPEN", "applicationStatus": "INFO_REQUESTED" }
```

#### `POST /bank/applications/{id}/decision`
Ручное решение сотрудника. Модалка «Иные условия» заполняет `amount`, `term`, `productCode`.

Заголовок: `Idempotency-Key`. Запрос:
```json
{
  "type": "COUNTER",
  "amount": 1200000,
  "term": 24,
  "productCode": "WORKING",
  "rate": 21.5,
  "reason": "Сумма снижена по уровню долговой нагрузки"
}
```
`type`: `APPROVE`, `COUNTER`, `REJECT`. Для `REJECT` обязателен только `reason`.

Ответ `200`:
```json
{ "status": "COUNTER_OFFER", "decisionId": "dec-31", "mode": "MANUAL", "decidedBy": "u-9", "decidedAt": "2026-10-20T13:10:00Z" }
```
Ошибка: `422`, если условия вне рамок продукта.

#### `GET /bank/applications/{id}/history`
История статусов и действий: кто, когда, что.

Ответ `200`:
```json
[
  { "at": "2026-10-20T10:25:00Z", "actor": { "type": "CLIENT", "name": "Петрова А. С." }, "action": "SUBMIT", "from": "DRAFT", "to": "SUBMITTED" },
  { "at": "2026-10-20T10:26:00Z", "actor": { "type": "SYSTEM" }, "action": "CHECKS_DONE", "from": "SUBMITTED", "to": "DATA_CHECK" },
  { "at": "2026-10-20T13:10:00Z", "actor": { "type": "OFFICER", "name": "Иванова Е. П." }, "action": "DECISION", "from": "ON_DECISION", "to": "COUNTER_OFFER", "comment": "Сумма снижена по уровню долговой нагрузки" }
]
```

### 4.13 Технические

| Метод | Путь | Что делает |
|---|---|---|
| GET | `/api/docs` | Swagger UI (стр. 7) |
| GET | `/api/openapi.json` | Спецификация OpenAPI (стр. 7) |
| GET | `/health` | `{ "status": "UP" }` — процесс жив (для Docker healthcheck) |
| GET | `/ready` | `{ "status": "UP", "db": "UP", "mocks": "UP" }` — готов принимать запросы |
| POST | `/admin/seed/reset` | Сброс демо-данных, только в demo-профиле. Ответ `204`. Доп. |

### 4.14 Моки внешних сервисов (отдельный контейнер)

Вызываются только из зоны Integrations, не из UI. Поведение мока задаётся ИНН из seed: по одному ИНН на каждую ветку «Схемы решения».

| Путь мока | Источник | Основа | Ответ (суть) |
|---|---|---|---|
| `GET /mock/core-banking/clients/{inn}` | АБС банка: профиль клиента | Уточн. | `{ inn, name, type, clientSince, regAddress, actualAddress, taxRegime, persons[{ personId, fullName, inn, roles[HEAD/FOUNDER/REPRESENTATIVE], share?, mchdNumber? }] }` |
| `GET /mock/core-banking/turnover/{inn}` | АБС банка: обороты по счетам | Уточн. | `{ avgMonthlyIncome, months[{ month, income, outcome }] }` |
| `GET /mock/egrul/{inn}` | ЕГРЮЛ (ООО) | ТЗ, стр. 13; путь — Доп. | `{ inn, name, ogrn, state, regDate, regAddress, okved, head, founders[] }` — пример ниже |
| `GET /mock/egrip/{inn}` | ЕГРИП (ИП) | ТЗ, стр. 13; путь — Доп. | `{ inn, ogrnip, fullName, state, regDate, okved }` |
| `GET /mock/transparent-business/{inn}` | «Прозрачный бизнес» ФНС | ТЗ, стр. 13 | `{ massAddress, massHead, taxRegime, employees }` |
| `POST /mock/bki/report` | БКИ | ТЗ, стр. 13 | запрос `{ inn, consentId }` → `{ loans[{ amount, balance, overdueDays }], totalDebt, maxOverdueDays }` |
| `GET /mock/fns/debts/{inn}` | Налоговая задолженность | Доп. | `{ hasDebt, amount }` |
| `GET /mock/girbo/{inn}` | ГИР БО (только ООО) | Доп. | `{ year, revenue, netProfit }` |
| `GET /mock/courts/{inn}` | КАД / ФССП / ЕФРСБ | Доп. | `{ arbitration[{ caseNumber, role, amount }], enforcement[{ number, amount }], bankruptcy }`; `role`: `DEFENDANT` / `PLAINTIFF` |
| `GET /mock/mchd/{number}` | Реестр МЧД (номер берётся из АБС, не от пользователя) | Доп. | `{ valid, principalInn, representative{ fullName, inn }, validUntil, powers[] }` |

Неизвестный мокам ИНН или номер МЧД, а также ГИР БО для ИП — `404 { code: "NOT_FOUND", message }` (Доп.).

Связь пользователя с данными банка: у `User` есть `personInn`. Полномочия определяются так:
1. Если среди `persons` клиента есть запись с этим ИНН и ролью `HEAD` — это руководитель.
2. Если есть роль `REPRESENTATIVE` с `mchdNumber` — доверенность проверяется в реестре МЧД.
3. Иначе полномочий нет.

Пример — ЕГРЮЛ:
```json
{
  "inn": "7712345678",
  "name": "ООО «Северная пекарня»",
  "ogrn": "1157746000000",
  "state": "ACTIVE",
  "regDate": "2015-03-12",
  "regAddress": "г. Москва, ул. Пекарская, 7",
  "okved": { "code": "10.71", "name": "Производство хлеба" },
  "head": { "fullName": "Смирнов Олег Викторович", "inn": "771234567890" },
  "founders": [{ "fullName": "Смирнов Олег Викторович", "share": 100 }]
}
```

---

## 5. Модель данных (кратко)

| Сущность | Ключевые поля | Зона |
|---|---|---|
| `User` | id, login, passwordHash, role, companyId, personInn, fullName, position | Identity |
| `Authority` | applicationId, type (HEAD/MCHD), mchdNumber, validUntil, verifiedAt — снимок при создании черновика; определяет Company по данным банка | Origination |
| `Product` | code, type, goals[], amountMin/Max, termMin/Max, rateMin/Max, minBusinessAgeMonths, guarantorForLlc, graceMaxMonths | Catalog |
| `Company` | inn, ogrn, type (LLC/IE), name, state, regDate, regAddress, actualAddress, okved, taxRegime, clientSince | Company |
| `Person` | personId, companyInn, fullName, inn, roles[] (HEAD/FOUNDER/REPRESENTATIVE), share, mchdNumber | Company |
| `Turnover` | companyInn, avgMonthlyIncome, periodMonths, months[] — по счетам в банке | Company |
| `Application` | id, number, companyId, applicantId, productCode, amount, term, goal, purposeText, graceMonths, guarantorPersonId, status, stage, version | Origination |
| `Consent` | applicationId, subject (APPLICANT/GUARANTOR), type (BKI), givenBy, givenAt | Origination |
| `Prelim` | applicationId, outcome, limit, prelimRate, amountFits, stopReason, choice (REDUCE_AMOUNT/KEEP_AMOUNT), amount, term, productCode (параметры, по которым посчитан) | Scoring |
| `CheckResult` | applicationId, source, status, requestedAt, payload, summary | Integrations |
| `ScoreResult` | applicationId, score, riskBand, factors[{code, label, impact}], autoDecisionEligible, modelVersion | Scoring |
| `Decision` | id, applicationId, type, amount, term, productCode, rate, reason, mode (AUTO/MANUAL), officerId | Scoring / Underwriting |
| `InfoRequest` | id, applicationId, items[], comment, status, answers[] | Underwriting |
| `Offer` / `Contract` | applicationId, terms, validUntil, acceptedAt, signedAt, disbursedAt | Offer |
| `Loan` / `ScheduleItem` | loanId, date, principal, interest, total, status | Servicing |
| `StatusEvent` / `AuditLog` | applicationId, from, to, actor, action, at, comment | Сквозное |
| `Notification` | id, userId, type, title, text, link, read | Сквозное |
| `Ticket` | id, userId, applicationId?, subject, text, status, reply | Support |

---

## 6. Что обязательно сдать по API (стр. 7)

- OpenAPI/Swagger (`/api/docs`) и описание API — этот документ можно положить в репозиторий как `docs/API.md`.
- Коллекция запросов (Postman / Bruno / .http) по сквозному сценарию: вход → каталог → черновик → этап 1 → этап 2 → отправка → разбор сотрудником → решение → принятие оффера → подписание → выдача.
- Диаграмма состояний заявки (раздел 2).
- Модель данных (раздел 5).
- Тестовые сценарии и синтетические данные: компании-клиенты банка под каждую ветку «Схемы решения». Ветки: одобрение, стоп-отказ, сумма выше лимита, иные условия, запрос информации, заявитель по МЧД, МЧД недействительна (нет полномочий). Матрица сценариев (логин → компания → ветка) — `backend/SCENARIOS.md`.
- THIRD_PARTY_NOTICES со всеми моками из 4.14.

---

## 7. Вопросы к организаторам

ТЗ этого не определяет:

1. Достаточно ли модульного монолита для критерия «Enterprise-ориентированность» (стр. 9, п. 05), или ждут отдельные сервисы?
2. Нужна ли в сдаче коллекция в конкретном формате (Postman), или подойдёт любой (стр. 7: «коллекция запросов»)?
3. Обязательна ли асинхронность обработки (очередь сообщений), или синхронный конвейер с моками допустим (стр. 9, п. 05 — отказоустойчивость, масштабирование)?

Внутренние открытые вопросы команды (решаем сами):

- Нужен ли отдельный статус между «Черновик» и «Подана» для результата этапа 1.
- Представитель по МЧД: второй аккаунт компании или роль внутри одного.
- Поддержка: отвечает сотрудник банка (нужен раздел в его кабинете) или только частые вопросы.
