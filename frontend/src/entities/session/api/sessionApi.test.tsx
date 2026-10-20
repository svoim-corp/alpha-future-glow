import { clientLogin, clientUser } from '@test/fixtures/session';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { useGetMeQuery, useLoginMutation, useLogoutMutation } from './sessionApi';

const Me = () => {
	const { data, isError } = useGetMeQuery();

	if (isError) {
		return <p role="alert">Не авторизован</p>;
	}

	return <p>{data ? data.fullName : 'Загрузка'}</p>;
};

const Actions = () => {
	const [login, loginState] = useLoginMutation();
	const [logout] = useLogoutMutation();

	return (
		<>
			<button type="button" onClick={() => void login({ login: 'smirnov', password: 'demo123' })}>Войти</button>
			<button type="button" onClick={() => void logout()}>Выйти</button>
			<p>{loginState.data ? `redirect:${loginState.data.redirect}` : ''}</p>
		</>
	);
};

describe('sessionApi', () => {
	it('getMe возвращает текущего пользователя', async () => {
		server.use(http.get(apiUrl('/auth/me'), () => HttpResponse.json(clientUser)));
		renderWithProviders(<Me />);

		expect(await screen.findByText('Смирнов Олег Викторович')).toBeInTheDocument();
	});

	it('getMe при 401 переходит в состояние ошибки', async () => {
		server.use(http.get(apiUrl('/auth/me'), () => HttpResponse.json({ code: 'UNAUTHORIZED', message: 'Нет сессии' }, { status: 401 })));
		renderWithProviders(<Me />);

		expect(await screen.findByRole('alert')).toHaveTextContent('Не авторизован');
	});

	it('login отправляет логин и пароль и возвращает адрес перехода', async () => {
		server.use(http.post(apiUrl('/auth/login'), async ({ request }) => {
			const body = await request.json();

			return body && typeof body === 'object' && 'login' in body && body.login === 'smirnov'
				? HttpResponse.json(clientLogin)
				: HttpResponse.json({ code: 'UNAUTHORIZED', message: 'Неверный логин или пароль' }, { status: 401 });
		}));
		const { user } = renderWithProviders(<Actions />);

		await user.click(screen.getByRole('button', { name: 'Войти' }));

		expect(await screen.findByText('redirect:/')).toBeInTheDocument();
	});

	it('logout завершает сессию', async () => {
		let called = false;
		server.use(http.post(apiUrl('/auth/logout'), () => {
			called = true;

			return new HttpResponse(null, { status: 204 });
		}));
		const { user } = renderWithProviders(<Actions />);

		await user.click(screen.getByRole('button', { name: 'Выйти' }));

		await waitFor(() => expect(called).toBe(true));
	});
});
