import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { Route, Routes } from 'react-router';

import { StartApplicationButton } from './StartApplicationButton';

const renderButton = (amount: number | null = 1500000) => renderWithProviders(
	<Routes>
		<Route
			path="/"
			element={<StartApplicationButton productCode="WORKING" amount={amount} term={24} goal="WORKING_CAPITAL" />}
		/>
		<Route path="/applications/:id/edit" element={<p>Мастер заявки</p>} />
	</Routes>,
);

describe('StartApplicationButton', () => {
	it('создаёт черновик с параметрами подбора и открывает мастер', async () => {
		let body: unknown;
		server.use(http.post(apiUrl('/applications'), async ({ request }) => {
			body = await request.json();

			return HttpResponse.json(
				{ id: 'a-1031', status: 'DRAFT', version: 1, redirect: '/applications/a-1031/edit' },
				{ status: 201 },
			);
		}));
		const { user } = renderButton();

		await user.click(screen.getByRole('button', { name: 'Оформить' }));

		expect(await screen.findByText('Мастер заявки')).toBeInTheDocument();
		expect(body).toEqual({ productCode: 'WORKING', amount: 1500000, term: 24, goal: 'WORKING_CAPITAL' });
	});

	it('показывает причину отказа, если сумма вне рамок продукта', async () => {
		server.use(http.post(apiUrl('/applications'), () => HttpResponse.json(
			{ code: 'BUSINESS_RULE', message: 'Сумма вне рамок продукта' },
			{ status: 422 },
		)));
		const { user } = renderButton();

		await user.click(screen.getByRole('button', { name: 'Оформить' }));

		expect(await screen.findByRole('alert')).toHaveTextContent('Сумма вне рамок продукта');
	});

	it('без суммы кнопка недоступна', () => {
		renderButton(null);

		expect(screen.getByRole('button', { name: 'Оформить' })).toBeDisabled();
	});
});
