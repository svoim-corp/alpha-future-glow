import { clientUser } from '@test/fixtures/session';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { ClientHeader } from './ClientHeader';

describe('ClientHeader', () => {
	it('показывает навигацию клиента, поддержку и имя пользователя', async () => {
		server.use(http.get(apiUrl('/auth/me'), () => HttpResponse.json(clientUser)));
		renderWithProviders(<ClientHeader />);

		const nav = screen.getByRole('navigation', { name: 'Основная навигация' });

		expect(nav).toHaveTextContent('Главная');
		expect(screen.getByRole('link', { name: 'Каталог' })).toHaveAttribute('href', '/products');
		expect(screen.getByRole('link', { name: 'Мои кредиты' })).toHaveAttribute('href', '/credits');
		expect(screen.getByRole('link', { name: 'Платёжный календарь' })).toHaveAttribute('href', '/credits/calendar');
		expect(screen.getByRole('link', { name: 'Поддержка' })).toHaveAttribute('href', '/support');
		expect(await screen.findByText('Смирнов Олег Викторович')).toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'Выйти' })).toBeInTheDocument();
	});

	it('текущий раздел помечен aria-current', () => {
		server.use(http.get(apiUrl('/auth/me'), () => HttpResponse.json(clientUser)));
		renderWithProviders(<ClientHeader />, { route: '/' });

		expect(screen.getByRole('link', { name: 'Главная' })).toHaveAttribute('aria-current', 'page');
	});
});
