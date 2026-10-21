import { infoRequestedApplication } from '@test/fixtures/applications';
import { loanDetailsFixture, loanFixture } from '@test/fixtures/loans';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen, within } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { MyLoans } from './MyLoans';

const list = (items: unknown[]) => HttpResponse.json({ items, page: 1, size: 20, total: items.length });

const approved = {
	...infoRequestedApplication,
	id: 'a-1021',
	number: 'КЗ-2026-001021',
	amount: 800000,
	term: 12,
	product: { code: 'EXPRESS', name: 'Экспресс-кредит' },
	status: 'APPROVED',
};

describe('MyLoans', () => {
	it('показывает обзор, ближайший платёж и график действующего кредита', async () => {
		server.use(
			http.get(apiUrl('/loans'), () => HttpResponse.json([loanFixture])),
			http.get(apiUrl('/loans/l-55'), () => HttpResponse.json(loanDetailsFixture)),
			http.get(apiUrl('/applications'), () => list([approved])),
		);
		renderWithProviders(<MyLoans />);

		expect(await screen.findByRole('heading', { name: 'Оборотный кредит' })).toBeInTheDocument();
		expect(screen.getByText('КД-2026-0055', { exact: false })).toBeInTheDocument();
		expect(screen.getByRole('tab', { name: /Действующие\s*1/ })).toHaveAttribute('aria-selected', 'true');
		expect(screen.getByRole('tab', { name: /Одобренные\s*1/ })).toBeInTheDocument();
		expect(screen.getByRole('heading', { name: '20 ноября' })).toBeInTheDocument();
		expect(within(screen.getByRole('table')).getAllByRole('row')).toHaveLength(3);
		expect(screen.getByText('Предстоит')).toBeInTheDocument();
		expect(screen.getByRole('link', { name: 'Платёжный календарь' })).toHaveAttribute('href', '/credits/calendar');
	});

	it('на вкладке «Одобренные» показывает предложение банка', async () => {
		server.use(
			http.get(apiUrl('/loans'), () => HttpResponse.json([loanFixture])),
			http.get(apiUrl('/loans/l-55'), () => HttpResponse.json(loanDetailsFixture)),
			http.get(apiUrl('/applications'), () => list([approved])),
			http.get(apiUrl('/applications/a-1021/offer'), () => HttpResponse.json({ code: 'NOT_FOUND', message: 'x' }, { status: 404 })),
		);
		const { user } = renderWithProviders(<MyLoans />);

		await user.click(await screen.findByRole('tab', { name: /Одобренные/ }));

		expect(screen.getByRole('heading', { name: 'Экспресс-кредит одобрен' })).toBeInTheDocument();
		expect(screen.getByRole('link', { name: /Посмотреть условия/ })).toHaveAttribute('href', '/applications/a-1021');
	});

	it('без кредитов предлагает подобрать финансирование', async () => {
		server.use(
			http.get(apiUrl('/loans'), () => HttpResponse.json([])),
			http.get(apiUrl('/applications'), () => list([])),
		);
		renderWithProviders(<MyLoans />);

		expect(await screen.findByText('Выданных кредитов пока нет')).toBeInTheDocument();
		expect(screen.getByRole('link', { name: 'Подобрать финансирование' })).toHaveAttribute('href', '/products');
	});

	it('при ошибке даёт повторить', async () => {
		server.use(
			http.get(apiUrl('/loans'), () => HttpResponse.json({ code: 'INTERNAL', message: 'x' }, { status: 500 }), { once: true }),
			http.get(apiUrl('/loans'), () => HttpResponse.json([loanFixture])),
			http.get(apiUrl('/loans/l-55'), () => HttpResponse.json(loanDetailsFixture)),
			http.get(apiUrl('/applications'), () => list([])),
		);
		const { user } = renderWithProviders(<MyLoans />);

		await user.click(await screen.findByRole('button', { name: 'Повторить' }));

		expect(await screen.findByRole('heading', { name: 'Оборотный кредит' })).toBeInTheDocument();
	});
});
