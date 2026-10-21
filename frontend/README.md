# Frontend — «Умный кредитный кабинет»

Веб-клиент: ЛК предпринимателя и кабинет сотрудника банка. React 19 + TypeScript, Redux Toolkit / RTK Query, React Router, архитектура Feature-Sliced Design.

## С чего начать

- [Контекст проекта и состояние реализации](docs/PROJECT_CONTEXT.md) — что за продукт, что уже есть, что осталось и где лежит код.
- [План приведения интерфейса к дизайну](docs/UI_ALIGNMENT_PLAN.md) — порядок исправления компонентов, оболочки и страниц, критерии готовности.
- [AGENTS.md](AGENTS.md) — правила архитектуры, UI и проверок.
- [Основная спецификация API](../docs/API_SPEC.md) и [дополнение для планера](../docs/API_SPEC_ADDITIONAL.md) — требования и контракты.
- [UI-kit](docs/ui-kit/README.md), [витрина компонентов](docs/ui-kit/index.html), [галерея макетов](docs/designs/index.html) и [описание макетов](docs/designs/README.md) — визуальная основа интерфейса.

Страницы и основные сценарии уже реализованы на MSW; перенос дизайна частичный, есть расхождения с актуальным API. Большинство бизнес-эндпоинтов сервера пока отсутствует. Наличие страницы в клиенте не означает готовность сквозного сценария с сервером.

## Требования

- Node.js 22 (`.nvmrc`), npm 10

## Запуск

```bash
cp .env.example .env
npm ci
npm run dev        # http://localhost:3000
npm run dev:mock   # то же, но без бэкенда: запросы к API обслуживают моки в браузере
```

Запросы на `/api` dev-сервер проксирует на `VITE_API_PROXY_TARGET` (по умолчанию `http://localhost:8080`).

## Переменные окружения

| Переменная | Где | По умолчанию | Назначение |
|---|---|---|---|
| `VITE_API_BASE_URL` | сборка | `/api/v1` | базовый URL API для RTK Query |
| `VITE_API_PROXY_TARGET` | dev | `http://localhost:8080` | куда Vite проксирует `/api` |
| `VITE_MOCK` | сборка / dev | не задана; `true` в `.env.mock` | включить MSW в браузере; используется в `dev:mock` |

## Скрипты

| Команда | Что делает |
|---|---|
| `npm run dev` | dev-сервер |
| `npm run dev:mock` | dev-сервер с моками API в браузере (MSW, `VITE_MOCK=true`), бэкенд не нужен |
| `npm run build` | проверка типов + сборка в `dist/` |
| `npm run check` | typecheck + ESLint + Steiger (FSD) + тесты — обязателен перед коммитом |
| `npm run lint` / `lint:fix` | ESLint (включая границы слоёв FSD) |
| `npm run lint:fsd` | Steiger — структурный линтер FSD |
| `npm test` / `test:watch` / `test:coverage` | Vitest |

## Структура

```
src/
  app/        entrypoint, model (стор), routes, styles, mocks (демо API)
  pages/      главная, вход, каталог, заявки, кредиты, планер, поддержка, банк, 404
  widgets/    оболочка клиента, каталог, мастер заявки, трекер, календарь, очередь банка
  features/   вход/выход, создание и отправка заявки, оффер, подписание, планы, решение банка
  entities/   session, product, application, contract, loan, notification, plan, support
  shared/     api (baseApi RTK Query), config, lib, ui (переиспользуемые компоненты)
test/         setup, renderWithProviders, MSW-сервер и фикстуры
docs/         PROJECT_CONTEXT.md, ui-kit/, designs/
```

Библиотека документов и загрузка выписки исключены актуальным контрактом и удалены из кода. Ввод ИНН сохраняется до отдельной задачи.

В составе всего проекта клиент собирается `Dockerfile` и раздаётся nginx (`nginx.conf`, `/api` → `backend:8080`), запуск — из `../deploy`.

## Как проверяется FSD

- **ESLint** (`no-restricted-imports`, сгенерировано в `eslint.config.js`): импорт только в нижние слои, без кросс-импортов внутри слоя, только через public API, без `../../` между слайсами, `@test` — только в тестах.
- **Steiger** (`steiger.config.ts`): структура слоёв и сегментов, наличие public API, запрещённые имена сегментов.

## Тесты

Vitest + Testing Library + MSW. Сеть мокается на уровне HTTP (MSW), стор настоящий. Пороги покрытия: 80 % строк/функций/выражений, 70 % ветвлений.

## Демо без бэкенда

`npm run dev:mock` (режим `mock`, `.env.mock`) запускает Service Worker из MSW: все запросы `/api/v1/*`
обслуживаются моками из `src/app/mocks`. Состояние хранится в памяти вкладки — созданные заявки, планы
и решения банка живут до перезагрузки. Вход сохраняется на время сессии вкладки.

- логины: `smirnov` (клиент), `ivanova` (сотрудник банка); пароль у всех `demo123`;
- поданная клиентом заявка появляется в очереди банка, а решение сотрудника — у клиента в заявке;
- исходные фикстуры находятся в `src/app/mocks/fixtures`, тестовые — в `test/fixtures`; это отдельные копии, их нужно согласованно обновлять. Целевые контракты — `../docs/API_SPEC.md` и `../docs/API_SPEC_ADDITIONAL.md`; известные отставания моков описаны в контексте проекта;
- Service Worker (`public/mockServiceWorker.js`) создан `npx msw init public`; в обычных режимах не подключается.

## Просмотр дизайнов

Из директории `frontend/`:

```bash
python3 -m http.server 8766 --bind 127.0.0.1 --directory docs
```

Галерея: [http://127.0.0.1:8766/designs/](http://127.0.0.1:8766/designs/), UI-kit: [http://127.0.0.1:8766/ui-kit/](http://127.0.0.1:8766/ui-kit/). HTML также открывается напрямую, сборка не нужна.

Известные ошибки проверок на дату аудита — в [контексте проекта](docs/PROJECT_CONTEXT.md#проверки).
