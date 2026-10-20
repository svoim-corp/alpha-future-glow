import { clientUser, officerUser } from '@test/fixtures/session';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { Route, Routes } from 'react-router';

import { LoginPage } from './LoginPage';

const renderPage = () => renderWithProviders(
	<Routes>
		<Route path="/login" element={<LoginPage />} />
		<Route path="/" element={<p>Главная</p>} />
		<Route path="/bank/queue" element={<p>Очередь</p>} />
	</Routes>,
	{ route: '/login' },
);

describe('LoginPage', () => {
	it('для неавторизованного показывает форму и демо-учётки', async () => {
		server.use(http.get(apiUrl('/auth/me'), () => HttpResponse.json({ code: 'UNAUTHORIZED', message: 'Нет сессии' }, { status: 401 })));
		renderPage();

		expect(await screen.findByRole('button', { name: 'Войти' })).toBeInTheDocument();
		expect(screen.getByRole('heading', { level: 1, name: 'Умный кредитный кабинет' })).toBeInTheDocument();
		expect(screen.getByText(/smirnov/)).toBeInTheDocument();
	});

	it('уже вошедшего клиента отправляет на главную', async () => {
		server.use(http.get(apiUrl('/auth/me'), () => HttpResponse.json(clientUser)));
		renderPage();

		expect(await screen.findByText('Главная')).toBeInTheDocument();
	});

	it('уже вошедшего сотрудника отправляет в очередь заявок', async () => {
		server.use(http.get(apiUrl('/auth/me'), () => HttpResponse.json(officerUser)));
		renderPage();

		expect(await screen.findByText('Очередь')).toBeInTheDocument();
	});
});
