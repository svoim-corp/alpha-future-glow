import { draftApplication } from '@test/fixtures/applications';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { HomePage } from './HomePage';

describe('HomePage', () => {
	it('собирает заявки, блок кредитов и кнопку «Получить финансирование»', async () => {
		server.use(
			http.get(apiUrl('/applications'), () => HttpResponse.json({ items: [draftApplication], page: 1, size: 20, total: 1 })),
			http.get(apiUrl('/loans/summary'), () => HttpResponse.json({
				activeLoans: 1,
				nextPayment: { date: '2026-11-15', amount: 77800 },
			})),
		);
		renderWithProviders(<HomePage />);

		expect(screen.getByRole('heading', { level: 1, name: 'Мои заявки' })).toBeInTheDocument();
		expect(screen.getAllByRole('link', { name: 'Получить финансирование' })[0]).toHaveAttribute('href', '/products');
		expect(await screen.findByRole('heading', { name: 'Экспресс-кредит' })).toBeInTheDocument();
		expect(await screen.findByRole('heading', { name: 'Мои кредиты' })).toBeInTheDocument();
	});
});
