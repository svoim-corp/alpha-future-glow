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
);

describe('BankApplicationPage', () => {
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
