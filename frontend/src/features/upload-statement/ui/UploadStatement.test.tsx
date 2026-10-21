import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { UploadStatement } from './UploadStatement';

const txt = () => new File(['1CClientBankExchange'], 'vypiska.txt', { type: 'text/plain' });

describe('UploadStatement', () => {
	it('кнопка недоступна, пока файл не выбран', () => {
		renderWithProviders(<UploadStatement applicationId="a-1" />);

		expect(screen.getByRole('button', { name: 'Загрузить выписку' })).toBeDisabled();
	});

	it('загружает выписку и показывает итоги', async () => {
		const spy = vi.fn();
		server.use(http.post(apiUrl('/applications/a-1/statement'), async ({ request }) => {
			spy((await request.formData()).get('file'));

			return HttpResponse.json({
				parsed: true,
				period: { from: '2025-10-01', to: '2026-09-30' },
				operations: 1284,
				avgMonthlyIncome: 1910000,
				newLimit: 2865000,
				amountFits: true,
			});
		}));
		const { user } = renderWithProviders(<UploadStatement applicationId="a-1" />);

		await user.upload(screen.getByLabelText(/Выписка из 1С/), txt());
		await user.click(screen.getByRole('button', { name: 'Загрузить выписку' }));

		expect(await screen.findByText('Запрошенная сумма укладывается в новый лимит')).toBeInTheDocument();
		expect(screen.getByText('01.10.2025 — 30.09.2026')).toBeInTheDocument();
		expect(spy).toHaveBeenCalledTimes(1);
	});

	it('предупреждает, если сумма выше нового лимита', async () => {
		server.use(http.post(apiUrl('/applications/a-1/statement'), () => HttpResponse.json({
			parsed: true,
			period: { from: '2025-10-01', to: '2026-09-30' },
			operations: 10,
			avgMonthlyIncome: 100000,
			newLimit: 150000,
			amountFits: false,
		})));
		const { user } = renderWithProviders(<UploadStatement applicationId="a-1" />);

		await user.upload(screen.getByLabelText(/Выписка из 1С/), txt());
		await user.click(screen.getByRole('button', { name: 'Загрузить выписку' }));

		expect(await screen.findByText('Запрошенная сумма выше нового лимита')).toBeInTheDocument();
	});

	it('объясняет, что файл не в формате 1С', async () => {
		server.use(http.post(apiUrl('/applications/a-1/statement'), () => HttpResponse.json(
			{ code: 'STATEMENT_FORMAT', message: 'Неверный формат' },
			{ status: 400 },
		)));
		const { user } = renderWithProviders(<UploadStatement applicationId="a-1" />);

		await user.upload(screen.getByLabelText(/Выписка из 1С/), txt());
		await user.click(screen.getByRole('button', { name: 'Загрузить выписку' }));

		expect(await screen.findByText(/формата 1CClientBankExchange с ИНН вашей компании/)).toBeInTheDocument();
	});
});
