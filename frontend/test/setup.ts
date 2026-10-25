import '@testing-library/jest-dom/vitest';

import { cleanup } from '@testing-library/react';

import { server } from './msw/server';

// jsdom не реализует top layer и наблюдение размеров; состояние меню проверяет React.
// Поведение настоящего popover дополнительно проверяется в браузере.
if (!HTMLElement.prototype.showPopover) {
	Object.defineProperties(HTMLElement.prototype, {
		showPopover: {
			configurable: true,
			value(this: HTMLElement) { this.style.display = 'block'; },
		},
		hidePopover: {
			configurable: true,
			value(this: HTMLElement) { this.style.removeProperty('display'); },
		},
	});
}
if (!globalThis.ResizeObserver) {
	vi.stubGlobal('ResizeObserver', class {
		observe = vi.fn();
		unobserve = vi.fn();
		disconnect = vi.fn();
	});
}

// Любой запрос без обработчика — ошибка теста: никаких «тихих» походов в сеть
beforeAll(() => server.listen({ onUnhandledFrame: 'error' }));
afterEach(() => {
	cleanup();
	server.resetHandlers();
});
afterAll(() => server.close());
