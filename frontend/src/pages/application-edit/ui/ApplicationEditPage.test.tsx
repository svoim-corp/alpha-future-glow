import { fullApplication } from '@test/fixtures/applications';
import { companyProfileFixture } from '@test/fixtures/companyProfile';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { Route, Routes } from 'react-router';

import { ApplicationEditPage } from './ApplicationEditPage';

const renderPage = () => renderWithProviders(
	<Routes>
		<Route path="/applications/:id/edit" element={<ApplicationEditPage />} />
		<Route path="/applications/:id" element={<p>Карточка заявки</p>} />
	</Routes>,
	{ route: '/applications/a-1031/edit' },
);

const calc = http.post(apiUrl('/calculator/payment'), () => HttpResponse.json({
	rate: { min: 18, max: 26 },
	payment: { min: 70000, max: 80000 },
	gracePayment: null,
	withinProductLimits: true,
}));

describe('ApplicationEditPage', () => {
	it('черновик на этапе 1 показывает компанию, параметры и предварительное решение', async () => {
		const stage1 = {
			...fullApplication,
			stage: 1 as const,
			prelim: null,
			locks: { paramsLocked: false, stage2Available: false },
		};
		server.use(http.get(apiUrl('/applications/a-1031'), () => HttpResponse.json(stage1)), calc);
		renderPage();

		expect(await screen.findByRole('heading', { name: 'Предварительное решение' })).toBeInTheDocument();
		expect(screen.getByLabelText('ИНН компании или ИП')).toBeInTheDocument();
		expect(screen.getByLabelText('Сумма, ₽')).toHaveValue('1500000');
		expect(screen.queryByRole('heading', { name: 'Полная заявка' })).not.toBeInTheDocument();
	});

	it('на этапе 2 показывает данные банка и полную заявку без выписки', async () => {
		server.use(
			http.get(apiUrl('/applications/a-1031'), () => HttpResponse.json(fullApplication)),
			http.get(apiUrl('/applications/a-1031/company-profile'), () => HttpResponse.json(companyProfileFixture)),
		);
		renderPage();

		expect(await screen.findByRole('heading', { name: 'Полная заявка' })).toBeInTheDocument();
		expect(await screen.findByText('г. Москва, ул. Пекарская, 7')).toBeInTheDocument();
		expect(screen.queryByLabelText(/Выписка/)).not.toBeInTheDocument();
		expect(screen.getByLabelText('Поручитель')).toBeInTheDocument();
	});

	it('можно вернуться с этапа 2 к этапу 1', async () => {
		server.use(
			http.get(apiUrl('/applications/a-1031'), () => HttpResponse.json(fullApplication)),
			http.get(apiUrl('/applications/a-1031/company-profile'), () => HttpResponse.json(companyProfileFixture)),
			calc,
		);
		const { user } = renderPage();

		await user.click(await screen.findByRole('button', { name: 'Вернуться к этапу 1' }));

		expect(screen.getByRole('heading', { name: 'Предварительное решение' })).toBeInTheDocument();
	});

	it('отправленная заявка перенаправляет в карточку', async () => {
		server.use(http.get(
			apiUrl('/applications/a-1031'),
			() => HttpResponse.json({ ...fullApplication, status: 'SUBMITTED' }),
		));
		renderPage();

		expect(await screen.findByText('Карточка заявки')).toBeInTheDocument();
	});

	it('при ошибке загрузки предлагает повторить', async () => {
		server.use(http.get(apiUrl('/applications/a-1031'), () => HttpResponse.json(
			{ code: 'INTERNAL', message: 'Ошибка' },
			{ status: 500 },
		)));
		renderPage();

		expect(await screen.findByText('Не удалось загрузить заявку')).toBeInTheDocument();
	});
});
