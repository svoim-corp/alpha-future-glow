import { clientUser, officerUser } from '@test/fixtures/session';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { createMemoryRouter, RouterProvider } from 'react-router';

import { routes } from './router';

const renderAt = (path: string) => {
	const router = createMemoryRouter(routes, { initialEntries: [path] });

	return { router, ...renderWithProviders(<RouterProvider router={router} />, { withRouter: false }) };
};

const asUser = (user: typeof clientUser ) => {
	server.use(http.get(apiUrl('/auth/me'), () => HttpResponse.json(user)));
};

const anonymous = () => {
	server.use(http.get(apiUrl('/auth/me'), () => HttpResponse.json(
		{ code: 'UNAUTHORIZED', message: 'Нет сессии' },
		{ status: 401 },
	)));
};

describe('маршруты приложения', () => {
	it('/ без сессии ведёт на страницу входа', async () => {
		anonymous();
		renderAt('/');

		expect(await screen.findByRole('heading', { level: 1, name: 'Умный кредитный кабинет' })).toBeInTheDocument();
	});

	it('/ для клиента — главная с шапкой клиента', async () => {
		asUser(clientUser);
		server.use(
			http.get(apiUrl('/applications'), () => HttpResponse.json({ items: [], page: 1, size: 20, total: 0 })),
			http.get(apiUrl('/loans/summary'), () => HttpResponse.json({ activeLoans: 0, nextPayment: null })),
		);
		renderAt('/');

		expect(await screen.findByRole('navigation', { name: 'Разделы кабинета' })).toBeInTheDocument();
	});

	it('сотрудник банка с адреса клиента попадает в очередь заявок', async () => {
		asUser(officerUser);
		const { router } = renderAt('/');

		await waitFor(() => expect(router.state.location.pathname).toBe('/bank/queue'));
	});

	it('неизвестный адрес — страница 404 со ссылкой на главную', () => {
		renderAt('/nope');

		expect(screen.getByRole('heading', { name: 'Страница не найдена' })).toBeInTheDocument();
		expect(screen.getByRole('link', { name: 'На главную' })).toHaveAttribute('href', '/');
	});

	it('при сбое проверки доступа предлагает повторить', async () => {
		server.use(http.get(apiUrl('/auth/me'), () => HttpResponse.json(
			{ code: 'UPSTREAM_UNAVAILABLE', message: 'down' },
			{ status: 502 },
		), { once: true }));
		renderAt('/');

		expect(await screen.findByRole('alert')).toHaveTextContent('Не удалось проверить доступ.');
	});
});
