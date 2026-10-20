import { fullApplication } from '@test/fixtures/applications';
import { counterOfferFixture, infoRequestsFixture, timelineFixture } from '@test/fixtures/tracking';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { Route, Routes } from 'react-router';

import { ApplicationPage } from './ApplicationPage';

const base = { ...fullApplication, number: 'КЗ-2026-001024' };
const renderPage = () => renderWithProviders(
	<Routes>
		<Route path="/applications/:id" element={<ApplicationPage />} />
		<Route path="/applications/:id/edit" element={<p>Мастер заявки</p>} />
	</Routes>,
	{ route: '/applications/a-1031' },
);
const app = (status: string) => http.get(apiUrl('/applications/a-1031'), () => HttpResponse.json({ ...base, status }));
const timeline = http.get(apiUrl('/applications/a-1031/timeline'), () => HttpResponse.json(timelineFixture));

describe('ApplicationPage', () => {
	it('при запросе банка показывает статус и форму ответа', async () => {
		server.use(
			app('INFO_REQUESTED'),
			timeline,
			http.get(apiUrl('/applications/a-1031/info-requests'), () => HttpResponse.json(infoRequestsFixture)),
		);
		renderPage();

		expect(await screen.findByText('Заявка КЗ-2026-001024')).toBeInTheDocument();
		expect(await screen.findByText('Банк проверил заявку и просит уточнить данные')).toBeInTheDocument();
		expect(await screen.findByLabelText('Адрес склада, где хранится товар')).toBeInTheDocument();
	});

	it('при иных условиях показывает предложение и кнопки решения', async () => {
		server.use(
			app('COUNTER_OFFER'),
			timeline,
			http.get(apiUrl('/applications/a-1031/offer'), () => HttpResponse.json(counterOfferFixture)),
		);
		renderPage();

		expect(await screen.findByRole('button', { name: 'Принять условия' })).toBeInTheDocument();
		expect(screen.getByRole('region', { name: 'Предложение банка' })).toHaveTextContent('1 200 000');
	});

	it('на подписании ведёт к странице подписания', async () => {
		server.use(
			app('SIGNING'),
			timeline,
			http.get(apiUrl('/applications/a-1031/offer'), () => HttpResponse.json(counterOfferFixture)),
		);
		renderPage();

		expect(await screen.findByRole('link', { name: 'Перейти к подписанию' }))
			.toHaveAttribute('href', '/applications/a-1031/sign');
		expect(screen.queryByRole('button', { name: 'Принять условия' })).not.toBeInTheDocument();
	});

	it('даёт перейти в поддержку с привязкой к заявке', async () => {
		server.use(app('SCORING'), timeline);
		renderPage();

		expect(await screen.findByRole('link', { name: 'Задать вопрос по заявке' }))
			.toHaveAttribute('href', '/support?applicationId=a-1031');
	});

	it('черновик отправляет в мастер', async () => {
		server.use(app('DRAFT'));
		renderPage();

		expect(await screen.findByText('Мастер заявки')).toBeInTheDocument();
	});
});
