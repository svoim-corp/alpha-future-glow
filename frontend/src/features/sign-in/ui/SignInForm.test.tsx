import { clientLogin, officerLogin } from '@test/fixtures/session';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { Route, Routes } from 'react-router';

import { SignInForm } from './SignInForm';

const renderForm = () => renderWithProviders(
	<Routes>
		<Route path="/" element={<SignInForm />} />
		<Route path="/target" element={<p>Главная клиента</p>} />
		<Route path="/bank/queue" element={<p>Очередь заявок</p>} />
	</Routes>,
);

describe('SignInForm', () => {
	it('без ввода показывает ошибки полей и не отправляет запрос', async () => {
		const { user } = renderForm();

		await user.click(screen.getByRole('button', { name: 'Войти' }));

		expect(await screen.findByText('Введите логин')).toBeInTheDocument();
		expect(screen.getByText('Введите пароль')).toBeInTheDocument();
	});

	it('сотрудник банка после входа попадает в очередь заявок', async () => {
		server.use(http.post(apiUrl('/auth/login'), () => HttpResponse.json(officerLogin)));
		const { user } = renderForm();

		await user.type(screen.getByLabelText('Логин'), 'ivanova');
		await user.type(screen.getByLabelText('Пароль'), 'demo123');
		await user.click(screen.getByRole('button', { name: 'Войти' }));

		expect(await screen.findByText('Очередь заявок')).toBeInTheDocument();
	});

	it('клиент после входа переходит на адрес из ответа', async () => {
		server.use(http.post(apiUrl('/auth/login'), () => HttpResponse.json({ ...clientLogin, redirect: '/target' })));
		const { user } = renderForm();

		await user.type(screen.getByLabelText('Логин'), 'smirnov');
		await user.type(screen.getByLabelText('Пароль'), 'demo123');
		await user.click(screen.getByRole('button', { name: 'Войти' }));

		expect(await screen.findByText('Главная клиента')).toBeInTheDocument();
	});

	it('при неверном пароле показывает понятную ошибку', async () => {
		server.use(http.post(apiUrl('/auth/login'), () => HttpResponse.json(
			{ code: 'UNAUTHORIZED', message: 'bad credentials' },
			{ status: 401 },
		)));
		const { user } = renderForm();

		await user.type(screen.getByLabelText('Логин'), 'smirnov');
		await user.type(screen.getByLabelText('Пароль'), 'wrong');
		await user.click(screen.getByRole('button', { name: 'Войти' }));

		expect(await screen.findByRole('alert')).toHaveTextContent('Неверный логин или пароль');
	});

	it('при сбое сервера предлагает повторить', async () => {
		server.use(http.post(apiUrl('/auth/login'), () => HttpResponse.json(
			{ code: 'UPSTREAM_UNAVAILABLE', message: 'down' },
			{ status: 502 },
		)));
		const { user } = renderForm();

		await user.type(screen.getByLabelText('Логин'), 'smirnov');
		await user.type(screen.getByLabelText('Пароль'), 'demo123');
		await user.click(screen.getByRole('button', { name: 'Войти' }));

		expect(await screen.findByRole('alert')).toHaveTextContent('Не удалось войти');
	});
});
