import { renderWithProviders } from '@test/render';
import { screen } from '@testing-library/react';
import { createMemoryRouter, RouterProvider } from 'react-router';

import { routes } from './router';

const renderAt = (path: string) => {
	const router = createMemoryRouter(routes, { initialEntries: [path] });

	return renderWithProviders(<RouterProvider router={router} />, { withRouter: false });
};

describe('маршруты приложения', () => {
	it('/ — главная страница', () => {
		renderAt('/');

		expect(screen.getByRole('heading', { level: 1, name: 'Кредитный кабинет' })).toBeInTheDocument();
	});

	it('неизвестный адрес — страница 404 со ссылкой на главную', () => {
		renderAt('/nope');

		expect(screen.getByRole('heading', { name: 'Страница не найдена' })).toBeInTheDocument();
		expect(screen.getByRole('link', { name: 'На главную' })).toHaveAttribute('href', '/');
	});
});
