import { setupWorker } from 'msw/browser';

import { handlers } from './handlers';

const worker = setupWorker(...handlers);

/** Запускает Service Worker с моками API: приложение работает без бэкенда */
export const startMockApi = () => worker.start({
	quiet: true,
	serviceWorker: { url: `${import.meta.env.BASE_URL}mockServiceWorker.js` },
});
