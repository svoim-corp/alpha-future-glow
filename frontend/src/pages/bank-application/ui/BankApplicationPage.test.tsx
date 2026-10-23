import {
	bankDetailsFixture,
	checksFixture,
	completenessFixture,
	historyFixture,
	scoreFixture,
} from '@test/fixtures/bank';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { Route, Routes } from 'react-router';

import { BankApplicationPage } from './BankApplicationPage';

const renderPage = () => renderWithProviders(
	<Routes><Route path="/bank/applications/:id" element={<BankApplicationPage />} /></Routes>,
	{ route: '/bank/applications/a-1024' },
);
const base = (status = 'ON_DECISION') => server.use(
	http.get(apiUrl('/bank/applications/a-1024'), () => HttpResponse.json({ ...bankDetailsFixture, status })),
	http.get(apiUrl('/bank/applications/a-1024/completeness'), () => HttpResponse.json(completenessFixture)),
	http.get(apiUrl('/bank/applications/a-1024/checks'), () => HttpResponse.json(checksFixture)),
	http.get(apiUrl('/bank/applications/a-1024/score'), () => HttpResponse.json(scoreFixture)),
	http.get(apiUrl('/bank/applications/a-1024/history'), () => HttpResponse.json(historyFixture)),
	http.get(apiUrl('/bank/applications/a-1024/info-requests'), () => HttpResponse.json([])),
);

describe('BankApplicationPage', () => {
	it('показывает запросы информации с ответами клиента, а без запросов не рисует секцию', async () => {
		base();
		server.use(http.get(apiUrl('/bank/applications/a-1024/info-requests'), () => HttpResponse.json([{
			id: 'r-1',
			status: 'ANSWERED',
			comment: 'Обороты расходятся с отчётностью',
			items: [{ key: 'turnover_note', type: 'TEXT', label: 'Почему выручка ниже оборотов' }],
			answers: [{ key: 'turnover_note', value: 'Возврат займа' }],
			createdAt: '2026-10-20T10:00:00Z',
			answeredAt: '2026-10-21T09:00:00Z',
		}])));
		renderPage();

		expect(await screen.findByRole('heading', { name: 'Запросы информации' })).toBeInTheDocument();
		expect(await screen.findByText('Возврат займа')).toBeInTheDocument();
	});

	it('без запросов информации секция не показывается', async () => {
		base();
		renderPage();

		await screen.findByText('Петрова Анна Сергеевна');

		expect(screen.queryByRole('heading', { name: 'Запросы информации' })).not.toBeInTheDocument();
	});

	it('собирает анкету, проверки, скоринг, решение и историю', async () => {
		base();
		renderPage();

		expect(await screen.findByText('Петрова Анна Сергеевна')).toBeInTheDocument();
		expect(await screen.findByText('Анкета заполнена полностью')).toBeInTheDocument();
		expect(await screen.findByText('ЕГРЮЛ')).toBeInTheDocument();
		expect(await screen.findByText('+80 баллов')).toBeInTheDocument();
		expect(await screen.findByText('Клиент Петрова А. С.: заявка подана')).toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'Иные условия' })).toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'Запросить информацию' })).toBeInTheDocument();
	});

	it('для уже решённой заявки панель решения не показывается', async () => {
		base('COUNTER_OFFER');
		renderPage();

		expect(await screen.findByText('Петрова Анна Сергеевна')).toBeInTheDocument();
		expect(screen.queryByRole('button', { name: 'Иные условия' })).not.toBeInTheDocument();
	});

	it('сбой скоринга не прячет остальные разделы', async () => {
		base();
		server.use(http.get(apiUrl('/bank/applications/a-1024/score'), () => HttpResponse.json(
			{ code: 'INTERNAL', message: 'x' },
			{ status: 500 },
		)));
		renderPage();

		expect(await screen.findByText('Не удалось загрузить раздел «Скоринг»')).toBeInTheDocument();
		expect(await screen.findByText('Петрова Анна Сергеевна')).toBeInTheDocument();
	});
});
