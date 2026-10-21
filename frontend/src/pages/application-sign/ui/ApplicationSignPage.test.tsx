import { contractFixture } from '@test/fixtures/contract';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { Route, Routes } from 'react-router';

import { ApplicationSignPage } from './ApplicationSignPage';

const renderPage = () => renderWithProviders(
	<Routes><Route path="/applications/:id/sign" element={<ApplicationSignPage />} /></Routes>,
	{ route: '/applications/a-1031/sign' },
);

describe('ApplicationSignPage', () => {
	it('показывает условия, документы и кнопку подписи', async () => {
		server.use(http.get(apiUrl('/applications/a-1031/contract'), () => HttpResponse.json(contractFixture)));
		renderPage();

		expect(await screen.findByText('Условия кредита')).toBeInTheDocument();
		expect(screen.getByRole('link', { name: 'Договор поручительства' })).toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'Подписать договор' })).toBeInTheDocument();
	});

	it('при ошибке предлагает повторить', async () => {
		server.use(http.get(apiUrl('/applications/a-1031/contract'), () => HttpResponse.json(
			{ code: 'NOT_FOUND', message: 'Нет' },
			{ status: 404 },
		)));
		renderPage();

		expect(await screen.findByText('Не удалось загрузить договор')).toBeInTheDocument();
	});
});
