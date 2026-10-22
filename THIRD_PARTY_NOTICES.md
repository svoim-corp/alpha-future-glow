# THIRD_PARTY_NOTICES

Сторонние сервисы, компоненты и библиотеки, которые использует «Умный кредитный кабинет» (ТЗ, стр. 7).
Собственный код проекта распространяется по лицензии MIT (`LICENSE`).

## 1. Внешние источники данных — только моки

Ни один реальный внешний сервис не вызывается. Все источники заменены моками: отдельный контейнер `mocks`
(`backend/cmd/mocks`, контракт — `docs/API_SPEC.md`, раздел 4.14). Данные моков синтетические и вымышленные:
ИНН, ОГРН, ФИО, адреса и номера дел придуманы для демо и не относятся к реальным организациям и людям
(`backend/SCENARIOS.md`). Сервер обращается к мокам только из зоны Integrations по `MOCKS_BASE_URL`.

| Мок | Что имитирует | Реальный сервис-прототип | Основа |
|---|---|---|---|
| `GET /mock/core-banking/clients/{inn}` | Карточка клиента и его представители в АБС банка | Автоматизированная банковская система (внутренняя) | Уточн. ТЗ |
| `GET /mock/core-banking/turnover/{inn}` | Обороты по счетам клиента в банке | АБС банка (внутренняя) | Уточн. ТЗ |
| `GET /mock/core-banking/accounts/{inn}` | Расчётные счета клиента и остатки | АБС банка (внутренняя) | Доп. |
| `POST /mock/core-banking/debits`, `GET /mock/core-banking/debits/{debitId}` | Поручение на списание в погашение кредита и его результат | АБС банка (внутренняя) | Доп. |
| `GET /mock/egrul/{inn}` | Сведения о юрлице | ЕГРЮЛ, ФНС России | ТЗ, стр. 13 |
| `GET /mock/egrip/{inn}` | Сведения об индивидуальном предпринимателе | ЕГРИП, ФНС России | ТЗ, стр. 13 |
| `GET /mock/transparent-business/{inn}` | Массовый адрес и руководитель, налоговый режим, численность | Сервис «Прозрачный бизнес», ФНС России | ТЗ, стр. 13 |
| `POST /mock/bki/report` | Кредитная история по согласию субъекта | Бюро кредитных историй | ТЗ, стр. 13 |
| `GET /mock/fns/debts/{inn}` | Налоговая задолженность | ФНС России | Доп. |
| `GET /mock/girbo/{inn}` | Бухгалтерская отчётность (только ООО) | ГИР БО, ФНС России | Доп. |
| `GET /mock/courts/{inn}` | Арбитражные дела, исполнительные производства, банкротство | КАД (арбитражные суды), ФССП, ЕФРСБ | Доп. |
| `GET /mock/mchd/{number}` | Машиночитаемая доверенность | Реестр МЧД, ФНС России | Доп. |

Названия реальных сервисов приведены только чтобы пояснить, что имитирует мок; проект с ними не интегрирован.

## 2. Библиотеки сервера (Go, `backend/go.mod`)

Прямые зависимости. Транзитивные зависимости перечислены в `backend/go.sum`.

| Модуль | Версия | Лицензия | Зачем |
|---|---|---|---|
| `github.com/gin-gonic/gin` | v1.11.0 | MIT | HTTP-роутер |
| `github.com/go-playground/validator/v10` | v10.29.0 | MIT | Валидация тел запросов |
| `github.com/jackc/pgx/v5` | v5.11.0 | MIT | Драйвер PostgreSQL |
| `github.com/jmoiron/sqlx` | v1.4.0 | MIT | Работа с SQL без ORM |
| `github.com/pressly/goose/v3` | v3.26.0 | MIT | SQL-миграции |
| `github.com/swaggo/swag` | v1.16.6 | MIT | OpenAPI из аннотаций |
| `github.com/swaggo/gin-swagger` | v1.6.1 | MIT | Swagger UI в gin |
| `github.com/swaggo/files` | v1.0.1 | MIT | Статика Swagger UI; внутри — Swagger UI (`swagger-api/swagger-ui`), Apache-2.0 |
| `golang.org/x/crypto` | v0.45.0 | BSD-3-Clause | bcrypt для паролей |
| `github.com/go-pdf/fpdf` | v0.9.0 | MIT | PDF графика платежей (`GET /loans/{id}/schedule/document`) |

Шрифт DejaVu Sans Condensed (обычный и полужирный, `backend/internal/servicing/fonts`) встроен в сервер для кириллицы в PDF.
Взят из поставки `go-pdf/fpdf`; лицензия — DejaVu Fonts License (свободная, на основе лицензии Bitstream Vera), https://dejavu-fonts.github.io/License.html.

## 3. Среда запуска (`deploy/docker-compose.yml`, `Dockerfile`)

| Компонент | Образ | Лицензия |
|---|---|---|
| PostgreSQL | `postgres:17-alpine` | PostgreSQL License |
| Go (сборка сервера и моков) | `golang:1.25-alpine` | BSD-3-Clause |
| Alpine Linux (образ сервера и моков) | `alpine:3.22` | MIT и лицензии пакетов Alpine |
| Node.js (сборка клиента) | `node:22-alpine` | MIT |
| nginx (раздача клиента) | `nginx:1.27-alpine` | BSD-2-Clause |

## 4. Библиотеки клиента

Зависимости клиента (`frontend/package.json`) дополняет команда фронтенда.

| Компонент | Версия | Лицензия | Зачем |
|---|---|---|---|
| [`@daypicker/react`](https://daypicker.dev/) | 10.0.2 | MIT | Переиспользуемый выбор даты с русской локалью, клавиатурной навигацией и оформлением токенами проекта; заменяет браузерные поля даты в планере |

Транзитивные зависимости календаря, включая совместимый пакет `react-day-picker`, зафиксированы в `frontend/package-lock.json`.
