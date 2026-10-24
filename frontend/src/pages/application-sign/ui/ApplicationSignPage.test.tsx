import { renderWithProviders } from '@test/render';
import { screen } from '@testing-library/react';
import { Route, Routes } from 'react-router';

import { ApplicationSignPage } from './ApplicationSignPage';

describe('ApplicationSignPage', () => {
	it('старый адрес подписания ведёт на страницу заявки', async () => {
		renderWithProviders(
			<Routes>
				<Route path="/applications/:id/sign" element={<ApplicationSignPage />} />
				<Route path="/applications/:id" element={<p>Страница заявки</p>} />
			</Routes>,
			{ route: '/applications/a-1031/sign' },
		);

		expect(await screen.findByText('Страница заявки')).toBeInTheDocument();
	});
});
