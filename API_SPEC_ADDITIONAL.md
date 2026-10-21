# Дополнение к API_SPEC: Документы и Планер

Версия: 20.10.2026. **Все контракты ниже — предложение фронтенда по макетам** `frontend/docs/designs/documents.html`, `planner.html`, `PLANNER.md`. В исходном ТЗ этих разделов нет, в основной спецификации (`API_SPEC.md`) их тоже нет — всё это **Доп.** (наше допущение). Бэкенд ещё не реализован; пока фронтенд работает на моках (MSW), которые повторяют этот контракт.

Соглашения те же, что в `API_SPEC.md` (раздел 3): префикс `/api/v1`, cookie-сессия, деньги — целые рубли, даты — ISO 8601, единый формат ошибки `{ code, message, fields?, traceId }`, `Idempotency-Key` для небезопасных повторов. Роль — `CLIENT`; данные видны только своей компании.

Новые доменные зоны:

| Зона | За что отвечает | Владеет данными |
|---|---|---|
| **Documents** | Личная библиотека файлов компании для повторного использования в заявках | `Document` |
| **Planner** | Планы и идеи развития, по которым Catalog подбирает финансирование | `BusinessPlan`, `PlanRecommendation` |

Новые теги кэша на фронтенде: `Document`, `Plan`.

---

## 1. Документы — `/documents`

Зона: Documents. Макет: `documents.html`.

### Модель `Document`

```json
{
  "id": "doc-1",
  "name": "Реквизиты компании",
  "category": "COMPANY",
  "extension": "PDF",
  "sizeBytes": 245760,
  "uploadedAt": "2026-10-12T09:00:00Z",
  "usage": { "applications": 2, "label": null }
}
```

- `category`: `COMPANY` (Компания), `STATEMENTS` (Выписки), `CONTRACTS` (Договоры), `PHOTOS` (Фото).
- `extension`: `PDF`, `TXT`, `JPG`, `PNG`.
- `usage.applications` — в скольких заявках файл уже использован; `usage.label` — подпись для договоров («Оборотный кредит»), иначе `null`.

### `GET /documents?category=COMPANY&q=реквизиты&sort=name`
Библиотека. Все параметры необязательны. `sort`: `name` (по названию, по умолчанию) или `-uploadedAt` (сначала новые).

Ответ `200`:
```json
{
  "items": [ { "...": "Document" } ],
  "counts": { "ALL": 6, "COMPANY": 2, "STATEMENTS": 1, "CONTRACTS": 2, "PHOTOS": 1 }
}
```
`counts` не зависит от `q` и `category` — по нему рисуются счётчики на категориях.

### `POST /documents`
Загрузка файла. `multipart/form-data`: поле `file` (`.pdf`, `.txt`, `.jpg`, `.png`, до 25 МБ) и поле `category`.

Ответ `201` — `Document`. Ошибки: `400 VALIDATION_ERROR` (`fields`: `file` — неверный тип или размер), `413 PAYLOAD_TOO_LARGE`.

### `GET /documents/{id}/download`
Скачивание или просмотр. Ответ — файл (`Content-Disposition: inline`). Для макета «Просмотр» фронтенд открывает ссылку в новой вкладке.

### `DELETE /documents/{id}`
Удаляет файл из библиотеки. Ответ `204`.

Копии, уже переданные банку в составе поданных заявок, **остаются** в истории заявки независимо от библиотеки (см. `PLANNER.md` / `PAGES.md`, раздел «Документы»).

### `POST /applications/{applicationId}/documents`
Кнопка «В заявку»: прикрепляет файл из библиотеки к черновику. Только `DRAFT` своей компании.

Запрос:
```json
{ "documentId": "doc-1" }
```
Ответ `204`. Ошибки: `404 NOT_FOUND`, `409 INVALID_TRANSITION` (заявка не черновик).

---

## 2. Планер — `/planner`

Зона: Planner. Макет: `planner.html`, правила — `frontend/docs/designs/PLANNER.md`.

### Модель `BusinessPlan`

```json
{
  "id": "p-1",
  "title": "Ремонт помещения",
  "note": "Обновить торговый зал и зону выпечки",
  "goal": "REAL_ESTATE",
  "budgetRub": 2400000,
  "dateMode": "PERIOD",
  "startDate": "2027-06-01",
  "endDate": "2027-08-31",
  "state": "ACTIVE",
  "notifications": { "inApp": true, "email": false },
  "recommendation": null,
  "createdAt": "2026-10-20T09:00:00Z"
}
```

- `dateMode`: `DAY` (один день, `endDate` = `startDate`), `PERIOD`, `UNSCHEDULED` (даты `null`).
- `state`: `ACTIVE`, `PAUSED`, `DONE`, `ARCHIVED`. В подборе участвует только `ACTIVE`.
- `goal` — код цели из `GET /products/goals` (Catalog).
- Подбор начинается с `startDate` (для `UNSCHEDULED` — сразу после сохранения); по окончании периода автоматическое продвижение прекращается. День начала определяется часовым поясом компании.
- `recommendation` — подбор по плану или `null`, если подходящего продукта нет:
```json
{
  "productCode": "INVEST",
  "productName": "Инвестиционный кредит",
  "reason": "Цель и сумма укладываются в рамки продукта",
  "availableFrom": "2026-10-20"
}
```
Это предварительное соответствие по параметрам, не решение банка и не гарантия одобрения.

### `GET /plans?state=ACTIVE`
Список планов компании. `state` необязателен (по умолчанию — все, кроме `ARCHIVED`).

Ответ `200`:
```json
{
  "items": [ { "...": "BusinessPlan" } ],
  "summary": { "total": 3, "withPeriod": 2, "unscheduled": 1, "budgetRub": 4250000 }
}
```
`budgetRub` — сумма ориентиров, а не кредитных обязательств.

### `POST /plans`
Запрос:
```json
{
  "title": "Новая печь для пекарни",
  "note": "Увеличить выпуск хлеба",
  "goal": "EQUIPMENT",
  "budgetRub": 1200000,
  "dateMode": "UNSCHEDULED",
  "startDate": null,
  "endDate": null,
  "notifications": { "inApp": true, "email": false }
}
```
Ответ `201` — `BusinessPlan`. Ошибки `400 VALIDATION_ERROR`: бюджет должен быть положительным, `endDate` не раньше `startDate`, для `DAY`/`PERIOD` дата начала обязательна. `email: true` требует отдельного согласия (в этой версии почта не отправляется — только флаг).

### `PATCH /plans/{id}`
Частичное изменение (в том числе `state`). Тело — подмножество полей `POST /plans` плюс `state`. Ответ `200` — `BusinessPlan`.

### `DELETE /plans/{id}`
Ответ `204`.

### Связь с уведомлениями
Когда по плану появляется подходящий вариант, создаётся уведомление `GET /notifications` с типом `PLAN_MATCH` (новый тип к `INFO_REQUESTED`, `DECISION`, `PAYMENT_DUE`, `TICKET_REPLY`), ссылка — `/planner`. Не чаще одного раза на один и тот же вариант.

### Виджет главной `plan-recommendations`
`GET /plans/recommendations` — идеи и планы, по которым уже есть подходящий продукт.

Ответ `200`:
```json
[
  {
    "planId": "p-3",
    "planTitle": "Новая печь для пекарни",
    "budgetRub": 1200000,
    "productCode": "INVEST",
    "productName": "Инвестиционный кредит",
    "reason": "Сумма и цель укладываются в рамки продукта"
  }
]
```
Пустой массив — виджет не показывается.

---

## 3. Вопросы организаторам

1. Нужны ли разделы «Документы» и «Планер» в рамках хакатона, или это расширение вне оценки?
2. Должна ли библиотека документов быть общей для клиентов одной компании или личной у пользователя?
3. Какие типы файлов и лимит размера допустимы для библиотеки (в макете — PDF/TXT/JPG/PNG, 25 МБ)?
4. Нужны ли реальные напоминания по почте для планов (в ТЗ уведомления только внутри продукта, стр. 11, п. 05)?
5. Как часто допустимо повторять подбор по плану (частота и срок актуальности)?
