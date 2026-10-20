# Backend — «Умный кредитный кабинет»

Go-сервер: REST API `/api/v1`, модульный монолит (пакет на доменную зону, схема БД на зону).
Контракт — `../API_SPEC.md`. Два бинаря: `cmd/api` (сервер) и `cmd/mocks` (моки внешних сервисов).

## Запуск

Через docker compose (БД, моки, сервер, клиент):

```bash
cd deploy && docker compose up --build
# только бэкенд с зависимостями:
docker compose up --build -d db mocks backend
```

Локально (нужен запущенный PostgreSQL 17, например `docker compose up -d db`):

```bash
cd backend
export DATABASE_URL='postgres://cabinet:cabinet@localhost:5432/cabinet?sslmode=disable'
make run                 # сервер на :8080
HTTP_ADDR=:8090 go run ./cmd/mocks
```

Swagger UI: http://localhost:8080/api/docs, спецификация: http://localhost:8080/api/openapi.json.

### Переменные окружения

| Переменная | По умолчанию | Назначение |
|---|---|---|
| `HTTP_ADDR` | `:8080` (моки — `:8090`) | адрес HTTP-сервера |
| `APP_PROFILE` | `dev` | `dev` / `demo` / `prod`; `demo` включает `POST /admin/seed/reset` |
| `DATABASE_URL` | — (обязательна) | строка подключения к PostgreSQL |
| `MOCKS_BASE_URL` | `http://localhost:8090` | адрес контейнера моков (для `/ready`, зона Integrations) |

### Тестовые учётки

Пароль у всех — `demo123`.

| Логин | Роль | Кто |
|---|---|---|
| `smirnov` | CLIENT | Смирнов О. В., генеральный директор, компания `c-1` |
| `petrova` | CLIENT | Петрова А. С., главный бухгалтер, компания `c-1` |
| `ivanova` | BANK_OFFICER | Иванова Е. П., кредитный инспектор |

## Makefile

| Цель | Что делает |
|---|---|
| `make build` / `vet` / `test` | `go build ./...`, `go vet ./...`, `go test ./...` |
| `make fmt` | `gofmt -l .` — список неотформатированных файлов (должен быть пустым) |
| `make run` | запуск сервера |
| `make swagger` | перегенерация `docs/` (нужен `swag`: `go install github.com/swaggo/swag/cmd/swag@latest`) |
| `make up` | `docker compose up --build` для всего проекта |

После изменения swag-аннотаций выполните `make swagger` и закоммитьте `docs/`.

## Зависимости

| Пакет | Зачем |
|---|---|
| `gin-gonic/gin` | HTTP-роутер (статический сегмент и параметр на одном уровне: `/products/goals`, `/products/:code`) |
| `jmoiron/sqlx` + `jackc/pgx/v5` | работа с PostgreSQL без ORM |
| `pressly/goose/v3` | SQL-миграции, применяются при старте, своя таблица версий на каждую схему |
| `swaggo/swag`, `gin-swagger`, `files` | OpenAPI из аннотаций и Swagger UI |
| `golang.org/x/crypto` | bcrypt для паролей |

Логи — стандартный `log/slog` (JSON). Версии зависимостей подобраны так, чтобы `go.mod` оставался на Go 1.25
(образ `golang:1.25-alpine` в `Dockerfile`); не поднимайте их через `go get -u` без обновления образа.

## Структура

```
cmd/api, cmd/mocks      точки входа
docs/                   сгенерированный OpenAPI (коммитится)
internal/
  platform/             сквозная инфраструктура
    config/             переменные окружения
    db/                 подключение, WithTx, Migrate (по схеме), db.JSON[T] для jsonb
    httpx/              формат ошибок, Fail, Bind, пагинация, If-Match, middleware, Routes
    auth/               middleware сессии, RequireAuth, RequireRole
    idempotency/        middleware для заголовка Idempotency-Key
    events/             синхронная шина доменных событий
  shared/               общие типы, интерфейсы между зонами (ports.go), события
  identity/             вход, сессии, роли (реализована)
  company/              минимальная таблица компаний (дальше расширяет зона)
  catalog, origination, integrations, scoring,
  underwriting, offer, servicing, support, notifications   заготовки
```

## Правила зон

- Пакеты зон **не импортируют друг друга**. Связь — через интерфейсы и типы из `internal/shared`;
  реализации подставляются в `cmd/api/main.go`. Зоны могут импортировать `platform/*` и `shared`.
- У зоны своя схема БД и миграции в `internal/<zone>/migrations/*.sql` (goose, `-- +goose Up`),
  встраиваются через `//go:embed migrations/*.sql` и применяются в `New` вызовом `db.Migrate(ctx, db, "<zone>", migrations)`.
  Образец — `internal/company`.
- Роуты регистрируются в `Register(r httpx.Routes)`: `r.Public` — без сессии, `r.Client` — с сессией,
  `r.Bank` — только `BANK_OFFICER` (префикс `/bank`).
- Ошибки в хендлерах — только через `httpx.Fail(c, err)`; прикладные ошибки создаются конструкторами
  `httpx.NotFound(...)`, `httpx.BusinessRule(...)` и т.д., остальное станет `500 INTERNAL`.
  Тело запроса — через `httpx.Bind` (поля с тегами `binding`).
- Статус заявки меняет только Origination (`shared.Applications.Transition`).
- Идемпотентность вешается на конкретный роут:

  ```go
  idem := idempotency.New(db)
  r.Client.POST("/applications/:id/submit", idem.Middleware(), h.submit)
  ```
- Публикуйте доменные события (`events.Bus.Publish`) только после коммита транзакции.
- Каждому хендлеру добавляйте swag-аннотации.

## Владельцы зон

| Разработчик | Зоны |
|---|---|
| A | identity, catalog, origination, offer, servicing, support, notifications |
| B | company, integrations, scoring, underwriting, бинарь моков (`cmd/mocks`) |
