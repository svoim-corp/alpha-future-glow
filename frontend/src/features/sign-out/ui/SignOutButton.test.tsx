import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { Route, Routes } from 'react-router';

import { SignOutButton } from './SignOutButton';

describe('SignOutButton', () => {
	it('завершает сессию и ведёт на страницу входа', async () => {
		let called = false;
		server.use(http.post(apiUrl('/auth/logout'), () => {
			called = true;

			return new HttpResponse(null, { status: 204 });
		}));
		const { user } = renderWithProviders(
			<Routes>
				<Route path="/" element={<SignOutButton />} />
				<Route path="/login" element={<p>Страница входа</p>} />
			</Routes>,
		);

		await user.click(screen.getByRole('button', { name: 'Выйти' }));

		expect(await screen.findByText('Страница входа')).toBeInTheDocument();
		expect(called).toBe(true);
	});
});
