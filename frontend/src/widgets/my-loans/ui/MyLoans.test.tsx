import { loanDetailsFixture, loanFixture } from '@test/fixtures/loans';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen, within } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { MyLoans } from './MyLoans';

describe('MyLoans', () => {
	it('показывает кредиты и открывает условия с графиком', async () => {
		server.use(
			http.get(apiUrl('/loans'), () => HttpResponse.json([loanFixture])),
			http.get(apiUrl('/loans/l-55'), () => HttpResponse.json(loanDetailsFixture)),
		);
		const { user } = renderWithProviders(<MyLoans />);

		await user.click(await screen.findByRole('button', { name: 'Условия и график: Оборотный кредит' }));
		const dialog = await screen.findByRole('dialog');

		expect(await within(dialog).findByText('КД-2026-0055')).toBeInTheDocument();
		expect(within(dialog).getAllByRole('row')).toHaveLength(3);
	});

	it('без кредитов предлагает подобрать финансирование', async () => {
		server.use(http.get(apiUrl('/loans'), () => HttpResponse.json([])));
		renderWithProviders(<MyLoans />);

		expect(await screen.findByText('Выданных кредитов пока нет')).toBeInTheDocument();
		expect(screen.getByRole('link', { name: 'Подобрать финансирование' })).toHaveAttribute('href', '/products');
	});

	it('при ошибке даёт повторить', async () => {
		server.use(
			http.get(apiUrl('/loans'), () => HttpResponse.json({ code: 'INTERNAL', message: 'x' }, { status: 500 }), { once: true }),
			http.get(apiUrl('/loans'), () => HttpResponse.json([loanFixture])),
		);
		const { user } = renderWithProviders(<MyLoans />);

		await user.click(await screen.findByRole('button', { name: 'Повторить' }));

		expect(await screen.findByRole('heading', { name: 'Оборотный кредит' })).toBeInTheDocument();
	});
});
