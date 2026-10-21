import { draftApplication, infoRequestedApplication } from '@test/fixtures/applications';
import { loanDetailsFixture, loanFixture } from '@test/fixtures/loans';
import { clientUser } from '@test/fixtures/session';
import { openInfoRequest } from '@test/fixtures/tracking';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { HomePage } from './HomePage';

const base = () => server.use(
	http.get(apiUrl('/auth/me'), () => HttpResponse.json(clientUser)),
	http.get(apiUrl('/applications'), () => HttpResponse.json({
		items: [infoRequestedApplication, draftApplication],
		page: 1,
		size: 20,
		total: 2,
	})),
	http.get(apiUrl('/applications/a-1024/info-requests'), () => HttpResponse.json([openInfoRequest])),
	http.get(apiUrl('/loans/summary'), () => HttpResponse.json({
		activeLoans: 1,
		nextPayment: { date: '2026-11-15', amount: 77800 },
	})),
	http.get(apiUrl('/loans'), () => HttpResponse.json([loanFixture])),
	http.get(apiUrl('/loans/l-55'), () => HttpResponse.json(loanDetailsFixture)),
	http.get(apiUrl('/documents'), () => HttpResponse.json({
		items: [{
			id: 'doc-1',
			name: 'Реквизиты компании',
			category: 'COMPANY',
			extension: 'PDF',
			sizeBytes: 1000,
			uploadedAt: '2026-10-12T09:00:00Z',
			usage: { applications: 0, label: null },
		}],
		counts: { ALL: 1, COMPANY: 1, STATEMENTS: 0, CONTRACTS: 0, PHOTOS: 0 },
	})),
	http.get(apiUrl('/plans/recommendations'), () => HttpResponse.json([])),
);

describe('HomePage', () => {
	it('приветствует по имени и показывает дату и компанию', async () => {
		base();
		renderWithProviders(<HomePage />);

		expect(await screen.findByRole('heading', { level: 1, name: /Олег$/ })).toBeInTheDocument();
		expect(screen.getByText(clientUser.company?.name ?? '')).toBeInTheDocument();
	});

	it('плитка заявок: этапы, запрос банка и черновик', async () => {
		base();
		renderWithProviders(<HomePage />);

		expect(await screen.findByText('1 500 000')).toBeInTheDocument();
		expect(screen.getByRole('list', { name: 'Этап рассмотрения' })).toBeInTheDocument();
		expect(await screen.findByText(openInfoRequest.items[0]?.label ?? '')).toBeInTheDocument();
		expect(screen.getByRole('link', { name: /Дополнить/ })).toHaveAttribute('href', '/applications/a-1024');
		expect(screen.getByRole('link', { name: /Продолжить/ })).toHaveAttribute('href', '/applications/a-1031/edit');
	});

	it('плитка платежа: дата, сумма и прогресс', async () => {
		base();
		renderWithProviders(<HomePage />);

		expect(await screen.findByText('77 800')).toBeInTheDocument();
		expect(await screen.findByRole('progressbar', { name: 'Выполнено платежей' })).toBeInTheDocument();
		expect(screen.getByRole('link', { name: /Посмотреть график/ })).toHaveAttribute('href', '/credits');
	});

	it('быстрые действия, каталог и документы ведут в разделы', async () => {
		base();
		renderWithProviders(<HomePage />);

		expect(await screen.findByRole('link', { name: 'Подать заявку' })).toHaveAttribute('href', '/products');
		expect(screen.getAllByRole('link', { name: 'Открыть планер' })[0]).toHaveAttribute('href', '/planner');
		expect(screen.getAllByRole('link', { name: /Подобрать финансирование/ })[0]).toHaveAttribute('href', '/products');
		expect(await screen.findByText('Реквизиты компании')).toBeInTheDocument();
	});

	it('без кредитов и заявок показывает пустые состояния', async () => {
		server.use(
			http.get(apiUrl('/auth/me'), () => HttpResponse.json(clientUser)),
			http.get(apiUrl('/applications'), () => HttpResponse.json({ items: [], page: 1, size: 20, total: 0 })),
			http.get(apiUrl('/loans/summary'), () => HttpResponse.json({ activeLoans: 0, nextPayment: null })),
			http.get(apiUrl('/loans'), () => HttpResponse.json([])),
			http.get(apiUrl('/documents'), () => HttpResponse.json({
				items: [],
				counts: { ALL: 0, COMPANY: 0, STATEMENTS: 0, CONTRACTS: 0, PHOTOS: 0 },
			})),
			http.get(apiUrl('/plans/recommendations'), () => HttpResponse.json([])),
		);
		renderWithProviders(<HomePage />);

		expect(await screen.findByText(/Отправленных заявок пока нет/)).toBeInTheDocument();
		expect(await screen.findByText(/Действующих кредитов пока нет/)).toBeInTheDocument();
		expect(await screen.findByText(/Документов пока нет/)).toBeInTheDocument();
	});
});
