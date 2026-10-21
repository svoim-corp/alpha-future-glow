import { goalsFixture, productsFixture } from '@test/fixtures/products';
import { clientUser } from '@test/fixtures/session';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { ProductsPage } from './ProductsPage';

describe('ProductsPage', () => {
	it('показывает заголовок страницы и каталог', async () => {
		server.use(
			http.get(apiUrl('/auth/me'), () => HttpResponse.json(clientUser)),
			http.get(apiUrl('/products/goals'), () => HttpResponse.json(goalsFixture)),
			http.get(apiUrl('/products'), () => HttpResponse.json(productsFixture)),
			http.post(apiUrl('/products/match'), () => HttpResponse.json([])),
			http.post(apiUrl('/calculator/payment'), () => HttpResponse.json({
				rate: { min: 1, max: 2 },
				payment: { min: 1, max: 2 },
				gracePayment: null,
				withinProductLimits: true,
			})),
		);
		renderWithProviders(<ProductsPage />);

		expect(screen.getByRole('heading', { level: 1, name: 'Каталог продуктов' })).toBeInTheDocument();
		expect(await screen.findByRole('button', { name: 'Все предложения' })).toBeInTheDocument();
		expect(await screen.findAllByRole('article')).toHaveLength(3);
	});
});
