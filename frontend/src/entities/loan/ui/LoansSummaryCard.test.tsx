import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { LoansSummaryCard } from './LoansSummaryCard';

describe('LoansSummaryCard', () => {
	it('показывает число кредитов, ближайший платёж и ссылку в «Мои кредиты»', async () => {
		server.use(http.get(apiUrl('/loans/summary'), () => HttpResponse.json({
			activeLoans: 1,
			nextPayment: { date: '2026-11-15', amount: 77800 },
		})));
		renderWithProviders(<LoansSummaryCard />);

		expect(await screen.findByRole('heading', { name: 'Мои кредиты' })).toBeInTheDocument();
		expect(screen.getByText(/Действующих кредитов: 1/)).toBeInTheDocument();
		expect(screen.getByText(/15\.11\.2026/)).toBeInTheDocument();
		expect(screen.getByRole('link', { name: 'Перейти к кредитам' })).toHaveAttribute('href', '/credits');
	});

	it('без выданных кредитов ничего не показывает', async () => {
		server.use(http.get(apiUrl('/loans/summary'), () => HttpResponse.json({ activeLoans: 0, nextPayment: null })));
		const { container } = renderWithProviders(<LoansSummaryCard />);

		await waitFor(() => expect(container).toBeEmptyDOMElement());
	});

	it('при ошибке предлагает повторить', async () => {
		server.use(http.get(apiUrl('/loans/summary'), () => HttpResponse.json(
			{ code: 'UPSTREAM_UNAVAILABLE', message: 'down' },
			{ status: 502 },
		), { once: true }));
		const { user } = renderWithProviders(<LoansSummaryCard />);

		expect(await screen.findByRole('alert')).toBeInTheDocument();
		server.use(http.get(apiUrl('/loans/summary'), () => HttpResponse.json({ activeLoans: 2, nextPayment: null })));
		await user.click(screen.getByRole('button', { name: 'Повторить' }));

		expect(await screen.findByText(/Действующих кредитов: 2/)).toBeInTheDocument();
	});
});
