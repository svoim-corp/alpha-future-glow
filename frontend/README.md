# Frontend — «Умный кредитный кабинет»

Веб-клиент: ЛК предпринимателя и кабинет сотрудника банка. React 19 + TypeScript, Redux Toolkit / RTK Query, React Router, архитектура Feature-Sliced Design.

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
  app/        entrypoint (main.tsx, провайдеры), model (стор), routes, styles
  pages/      home (пустая главная), not-found
  widgets/    — пусто
  features/   — пусто
  entities/   — пусто
  shared/     api (baseApi RTK Query), config, lib (format, store), ui (Button)
test/         setup, renderWithProviders, MSW-сервер и фикстуры
```

Правила архитектуры и тестирования — [AGENTS.md](./AGENTS.md). Контракт API — [`../API_SPEC.md`](../API_SPEC.md).

В составе всего проекта клиент собирается `Dockerfile` и раздаётся nginx (`nginx.conf`, `/api` → `backend:8080`), запуск — из `../deploy`.

## Как проверяется FSD

- **ESLint** (`no-restricted-imports`, сгенерировано в `eslint.config.js`): импорт только в нижние слои, без кросс-импортов внутри слоя, только через public API, без `../../` между слайсами, `@test` — только в тестах.
- **Steiger** (`steiger.config.ts`): структура слоёв и сегментов, наличие public API, запрещённые имена сегментов.

## Тесты

Vitest + Testing Library + MSW. Сеть мокается на уровне HTTP (MSW), стор настоящий. Пороги покрытия: 80 % строк/функций/выражений, 70 % ветвлений.

## Демо без бэкенда

`npm run dev:mock` (режим `mock`, `.env.mock`) запускает Service Worker из MSW: все запросы `/api/v1/*`
обслуживаются моками из `src/app/mocks`. Состояние хранится в памяти вкладки — созданные заявки, планы,
документы и решения банка живут до перезагрузки. Вход сохраняется на время сессии вкладки.

- логины: `smirnov` (клиент), `ivanova` (сотрудник банка); пароль у всех `demo123`;
- поданная клиентом заявка появляется в очереди банка, а решение сотрудника — у клиента в заявке;
- данные — копии `test/fixtures`; контракты — `API_SPEC.md` и `API_SPEC_ADDITIONAL.md`;
- Service Worker (`public/mockServiceWorker.js`) создан `npx msw init public`; в обычных режимах не подключается.
