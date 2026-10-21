# Разделение работы над бэкендом

Как два разработчика делят сервер, чтобы работать параллельно и не мешать друг другу.
Контракт API — `../API_SPEC.md` и `../API_SPEC_ADDITIONAL.md`; правила кода — `AGENTS.md`; демо-данные — `SCENARIOS.md`.

## Принцип

Делим по доменным зонам, а не по страницам. У каждого свои пакеты `internal/<zone>`, своя схема БД и свои эндпоинты.
Зоны общаются только через интерфейсы в `internal/shared`. Пока поставщик не написал реализацию, он отдаёт
заглушку из `ports.go` своей зоны, поэтому никто никого не ждёт.

| | Разработчик A — путь клиента | Разработчик B — данные банка, решение, кабинет банка |
|---|---|---|
| Зоны | identity, catalog, origination, offer, servicing, support, notifications | company, integrations, scoring, underwriting, planner |
| Бинари | `cmd/api` (общий) | `cmd/mocks` |
| Реализует для другого | `Products`, `Applications`, `Notifier` | `Companies`, `Authorities`, `Prelims`, `Decisions` |
| Материалы к сдаче | коллекция запросов по сквозному сценарию, диаграмма состояний | `THIRD_PARTY_NOTICES`, описание тестовых сценариев |

## Общие файлы — только по договорённости

| Файл | Правило |
|---|---|
| `internal/shared/*.go` | Меняет зона-поставщик интерфейса (указана в комментарии). Изменение сигнатуры — договориться и в том же коммите поправить реализацию, заглушку и всех потребителей |
| `cmd/api/main.go` | Каждый трогает только строки своих зон: `New`, `Wire`, список `zones` |
| `SCENARIOS.md` | A ведёт пользователей, B — компании и моки. Изменили строку — поправьте seed в том же коммите |
| `../API_SPEC.md` | Изменения контракта — отдельным коммитом, второй разработчик смотрит до мержа |
| `docs/` | Не мержить руками: при конфликте выполнить `make swagger` |
| `go.mod` | Без `go get -u` (см. README); новую зависимость обсудить заранее |

## Разработчик A

| # | Задача | Что сделать | Зависит от |
|---|---|---|---|
| A1 | **Catalog** | Схема и seed линейки продуктов (WORKING, INVEST, EXPRESS, OVERDRAFT…); `GET /products`, `GET /products/{code}`, `GET /products/goals`, `POST /products/match`, `POST /calculator/payment`. Заменить `productsStub` в `catalog/ports.go` | — |
| A2 | **Origination: черновик** | `POST /applications`: компания берётся из пользователя, полномочия — через `Authorities.Resolve`; если полномочий нет — `403`, иначе снимок полномочий сохраняется в заявке. Также `GET /applications?mine=true`, `GET /applications/{id}` (собирается из `Companies.Profile` и `Prelims`, вычисляет `prelimInvalidated`), `PATCH` (только `params` и `stage2.guarantor.personId`, `If-Match`), `DELETE`. Проверка доступа: только своя компания | A1; заглушки B |
| A3 | **Машина состояний** | Таблица переходов (раздел 2), `Transition` пишет `StatusEvent` и публикует `ApplicationStatusChanged`; `409 INVALID_TRANSITION`. Табличные тесты всех переходов. Заменить `applicationsStub` в `origination/ports.go` | A2 |
| A4 | **Согласия, проверка, отправка** | `POST .../consents/bki` (`403` без полномочия `BKI_CONSENT`), `POST .../guarantor/consent-bki`, `POST .../validate` (поручитель из `guarantorCandidates`, согласия, `truthConfirmed`), `POST .../submit` (`Idempotency-Key`, номер `КЗ-2026-…`, событие `ApplicationSubmitted`) | A3 |
| A5 | **Статус для клиента** | `GET .../timeline` (`whatsHappening`, `nextStep`, `actor`), `GET /bank/applications/{id}/history` | A3 |
| A6 | **Notifications** | `GET /notifications`, `POST /notifications/{id}/read`. Заменить `notifierStub` в `notifications/ports.go` | — |
| A7 | **Offer & Contract** | Подписка на `DecisionMade` → оффер из `Decisions.LatestDecision`; `GET .../offer`, `POST .../offer/accept`, `POST .../offer/decline`, `GET .../contract`, `GET .../contract/documents/{docId}` (печатная форма, договор поручительства только для ООО), `POST .../contract/sign`, `POST .../contract/disburse` (`Idempotency-Key`) | A3; `Decisions` от B |
| A8 | **Servicing** | Кредит и аннуитетный график при выдаче; `GET /loans/summary`, `GET /loans`, `GET /loans/{id}`, `GET /loans/schedule`, `GET /loans/schedule/upcoming`; напоминание `PAYMENT_DUE` | A7 |
| A9 | **Support** | `GET /support/faq` (seed вопросов), `POST /support/tickets`, `GET /support/tickets` | — |
| A10 | **Сдача** | Коллекция запросов: вход → каталог → черновик → этап 1 → этап 2 → отправка → разбор → решение → оффер → подписание → выдача. Диаграмма состояний | всё |

## Разработчик B

| # | Задача | Что сделать | Зависит от |
|---|---|---|---|
| B1 | **Моки** (`cmd/mocks`) | Все пути из `API_SPEC.md` 4.14, включая `core-banking/clients/{inn}` и `core-banking/turnover/{inn}`. Данные — строго по `SCENARIOS.md`, по одной компании на ветку решения | — |
| B2 | **Integrations** | HTTP-клиенты к мокам (таймауты, `502 UPSTREAM_UNAVAILABLE`), снимки ответов `CheckResult`. Адаптеры для Company и Scoring — через свои интерфейсы в `shared` (например, `shared/integrations.go`) | B1 |
| B3 | **Company** | Профиль клиента из АБС и реестров, кэш в своей схеме; `GET /applications/{id}/company-profile`. Заменить заглушку `Profile` в `company/ports.go` | B2 |
| B4 | **Полномочия** | `Authorities.Resolve`: роль человека в АБС (`HEAD` → руководитель), для `REPRESENTATIVE` — проверка МЧД в реестре, иначе `AuthorityNone` с причиной. Заменить `authoritiesStub` | B2 |
| B5 | **Prelim** | `POST /applications/{id}/prelim`: лимит по оборотам на счетах, стоп-причины (табличные тесты), проверка согласия через `Applications.BKIConsent`. `POST .../prelim/choice` (`REDUCE_AMOUNT` через `Applications.SetAmount`, `KEEP_AMOUNT`). Заменить `prelimsStub` | B3; `Products`, `Applications` от A |
| B6 | **Конвейер и скоринг** | Подписка на `ApplicationSubmitted`: проверки → `DATA_CHECK` → `SCORING` → балл и факторы → `ON_DECISION` → авторешение или ожидание сотрудника. Конвейер — в фоне с остановкой при завершении сервера, чтобы `submit` не ждал моки. `GET /bank/applications/{id}/score`. Заменить `decisionsStub`, публиковать `DecisionMade` | B5; A3 |
| B7 | **Очередь банка** | Своя таблица очереди, обновляемая по `ApplicationStatusChanged`; `GET /bank/applications/stats`, `GET /bank/applications` (фильтры, сортировка, пагинация), `POST /bank/applications/{id}/assign` (`409`, если взял другой) | A3 |
| B8 | **Разбор заявки** | `GET /bank/applications/{id}`, `GET .../completeness`, `GET .../checks`, `POST .../checks/refresh` | B3, B6 |
| B9 | **Запрос информации и решение** | `POST /bank/applications/{id}/info-requests` (→ `INFO_REQUESTED`, уведомление), клиентские `GET /applications/{id}/info-requests` и `POST .../answer` (→ `DATA_CHECK`), `POST /bank/applications/{id}/decision` (`Idempotency-Key`, `422` вне рамок продукта, `DecisionMade`) | B6; `Notifier` от A |
| B10 | **Planner** | `API_SPEC_ADDITIONAL.md`: `/plans`, `/plans/recommendations`, уведомления `PLAN_MATCH`. Делать последним: это «Доп.», у организаторов открыт вопрос, нужен ли раздел | `Products`, `Notifier` от A |
| B11 | **Сдача** | `THIRD_PARTY_NOTICES` со всеми моками, описание сценариев по `SCENARIOS.md` | B1 |

## Порядок и точки стыковки

```
A1 → A2 → A3 → A4 → A5          A6, A9 — в любой момент
                └──────→ A7 → A8
B1 → B2 → B3 → B5 → B6 → B8 → B9      B4 — после B2;  B7 — после A3;  B10 — в конце
```

| Стык | Кто | Что проверить вместе |
|---|---|---|
| 1. Создание черновика | A2 + B3, B4 | `smirnov` и `petrova` создают черновик, `volkov` получает `403` |
| 2. Этап 1 в мастере | A2 + B5 | `prelim` по всем стоп-сценариям и `sokolova` (сумма выше лимита) |
| 3. Отправка → конвейер | A4 + B6, B7 | заявка после `submit` проходит статусы и появляется в очереди |
| 4. Решение → оффер | B9 + A7 | `APPROVED`, `COUNTER_OFFER` (`morozov`), `REJECTED` (`polyakov`), `INFO_REQUESTED` (`kozlov`) |
| 5. Сквозной прогон | оба | каждый логин из `SCENARIOS.md` через `deploy/` до своего исхода, у `smirnov` и `kuznetsov` — до выдачи |

## Как работаем

- Ветка на задачу от `main` (`feature/a2-draft`, `feature/b5-prelim`), коммиты по правилам корневого `AGENTS.md`.
- Перед мержем: `make fmt vet build test`, swag-аннотации и `make swagger`, чек-лист из `AGENTS.md`.
- Заглушку заменяет только поставщик. Потребитель, которому нужно поведение раньше, договаривается о нём, а не пишет обход.
- Нужен новый интерфейс между A и B — сначала сигнатура и заглушка отдельным маленьким коммитом, потом реализация.

## Открытые вопросы

- Ответ на обращение в поддержку: сотрудник банка (тогда нужен раздел в его кабинете и `TICKET_REPLY`) или только FAQ.
- Каталог открыт без входа (`r.Public`) — подтвердить с фронтом.
- Нужен ли Planner для сдачи (вопрос организаторам в `API_SPEC_ADDITIONAL.md`).
