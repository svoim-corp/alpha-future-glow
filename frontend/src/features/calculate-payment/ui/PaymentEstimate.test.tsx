import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { PaymentEstimate } from './PaymentEstimate';

const params = { productCode: 'WORKING', amount: 1500000, term: 24 };
const normalize = (value: string | null) => (value ?? '').replace(/\s/gu, ' ');

describe('PaymentEstimate', () => {
	it('без параметров просит указать сумму и срок и не делает запрос', () => {
		renderWithProviders(<PaymentEstimate params={null} />);

		expect(screen.getByText('Укажите сумму и срок, чтобы увидеть платёж.')).toBeInTheDocument();
	});

	it('показывает платёж диапазоном и пояснение про ставку', async () => {
		server.use(http.post(apiUrl('/calculator/payment'), () => HttpResponse.json({
			rate: { min: 20, max: 28 },
			payment: { min: 76300, max: 82300 },
			gracePayment: null,
			withinProductLimits: true,
		})));
		renderWithProviders(<PaymentEstimate params={params} />);

		expect(normalize((await screen.findByText(/в месяц/)).textContent)).toBe('≈ 76 300 – 82 300 ₽ в месяц');
		expect(normalize(screen.getByText(/Ставка/).textContent)).toContain('20–28 % годовых');
	});

	it('при отсрочке показывает платёж только по процентам', async () => {
		server.use(http.post(apiUrl('/calculator/payment'), () => HttpResponse.json({
			rate: { min: 18, max: 26 },
			payment: { min: 162900, max: 189600 },
			gracePayment: { min: 90000, max: 130000, months: 6 },
			withinProductLimits: true,
		})));
		renderWithProviders(<PaymentEstimate params={{ ...params, graceMonths: 6 }} compact />);

		expect(await screen.findByText(/Первые 6 мес\./)).toBeInTheDocument();
		expect(screen.queryByText(/точную назовём/)).not.toBeInTheDocument();
	});

	it('вне рамок продукта подсказывает, что поменять', async () => {
		server.use(http.post(apiUrl('/calculator/payment'), () => HttpResponse.json({
			rate: { min: 20, max: 28 },
			payment: { min: 1, max: 2 },
			gracePayment: null,
			withinProductLimits: false,
		})));
		renderWithProviders(<PaymentEstimate params={params} />);

		expect(await screen.findByText('Сумма или срок вне рамок продукта')).toBeInTheDocument();
		expect(screen.getByText(/Измените сумму или срок/)).toBeInTheDocument();
	});

	it('при ошибке сервера показывает сообщение', async () => {
		server.use(http.post(apiUrl('/calculator/payment'), () => HttpResponse.json(
			{ code: 'BUSINESS_RULE', message: 'Продукт не найден' },
			{ status: 422 },
		)));
		renderWithProviders(<PaymentEstimate params={params} />);

		expect(await screen.findByRole('alert')).toHaveTextContent('Продукт не найден');
	});
});
