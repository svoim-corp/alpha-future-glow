import '@testing-library/jest-dom/vitest';

import { cleanup } from '@testing-library/react';

import { server } from './msw/server';

// Любой запрос без обработчика — ошибка теста: никаких «тихих» походов в сеть
beforeAll(() => server.listen({ onUnhandledFrame: 'error' }));
afterEach(() => {
	cleanup();
	server.resetHandlers();
});
afterAll(() => server.close());
