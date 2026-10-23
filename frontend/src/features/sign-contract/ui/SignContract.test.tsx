import { contractFixture } from '@test/fixtures/contract';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen, waitFor, within } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { SignContract } from './SignContract';

describe('SignContract', () => {
	it('подпись требует подтверждения в окне и уходит с ключом идемпотентности', async () => {
		const key = vi.fn();
		server.use(http.post(apiUrl('/applications/a-1/contract/sign'), ({ request }) => {
			key(request.headers.get('Idempotency-Key'));

			return HttpResponse.json({ signed: true, signedAt: '2026-10-20T14:00:00Z', signatureId: 'SIM-8F2A' });
		}), http.post(apiUrl('/applications/a-1/contract/disburse'), () => HttpResponse.json({
			status: 'ISSUED',
			loanId: 'l-55',
			disbursedAt: '2026-10-20T14:01:00Z',
			accountMasked: '•••• 4471',
		})));
		const { user } = renderWithProviders(<SignContract applicationId="a-1" contract={contractFixture} />);

		await user.click(screen.getByRole('button', { name: 'Подписать договор' }));
		expect(key).not.toHaveBeenCalled();
		await user.click(within(screen.getByRole('dialog')).getByRole('button', { name: 'Подтверждаю' }));

		await waitFor(() => expect(key).toHaveBeenCalledWith(expect.stringMatching(/^[0-9a-f-]{36}$/u)));
		await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
	});

	it('после подписи кредит выдаётся сразу, без отдельного шага «получить деньги»', async () => {
		const disburse = vi.fn();
		server.use(
			http.post(apiUrl('/applications/a-1/contract/sign'), () => HttpResponse.json(
				{ signed: true, signedAt: '2026-10-20T14:00:00Z', signatureId: 'SIM-8F2A' },
			)),
			http.post(apiUrl('/applications/a-1/contract/disburse'), () => {
				disburse();

				return HttpResponse.json({
					status: 'ISSUED',
					loanId: 'l-55',
					disbursedAt: '2026-10-20T14:01:00Z',
					accountMasked: '•••• 4471',
				});
			}),
		);
		const { user } = renderWithProviders(<SignContract applicationId="a-1" contract={contractFixture} />);

		await user.click(screen.getByRole('button', { name: 'Подписать договор' }));
		await user.click(within(screen.getByRole('dialog')).getByRole('button', { name: 'Подтверждаю' }));

		expect(await screen.findByText(/зачислены на счёт •••• 4471/)).toBeInTheDocument();
		expect(disburse).toHaveBeenCalledTimes(1);
		expect(screen.queryByRole('button', { name: 'Получить деньги' })).not.toBeInTheDocument();
		expect(screen.getByRole('link', { name: 'Перейти в «Мои кредиты»' })).toHaveAttribute('href', '/credits');
	});

	it('если выдача не прошла, договор остаётся подписанным и зачисление можно повторить', async () => {
		server.use(http.post(apiUrl('/applications/a-1/contract/disburse'), () => HttpResponse.json({
			status: 'ISSUED',
			loanId: 'l-55',
			disbursedAt: '2026-10-20T14:01:00Z',
			accountMasked: '•••• 4471',
		})));
		const signed = { ...contractFixture, signed: true };
		const { user } = renderWithProviders(<SignContract applicationId="a-1" contract={signed} />);

		expect(screen.getByText('Договор подписан')).toBeInTheDocument();
		await user.click(screen.getByRole('button', { name: 'Повторить зачисление' }));

		expect(await screen.findByText(/зачислены на счёт •••• 4471/)).toBeInTheDocument();
	});

	it('для уже выданного кредита сразу показывает итог', () => {
		const issued = { ...contractFixture, signed: true, disbursed: true };
		renderWithProviders(<SignContract applicationId="a-1" contract={issued} />);

		expect(screen.getByText('Кредит выдан')).toBeInTheDocument();
		expect(screen.queryByRole('button', { name: 'Получить деньги' })).not.toBeInTheDocument();
	});

	it('ошибка подписи показывается в окне', async () => {
		server.use(http.post(apiUrl('/applications/a-1/contract/sign'), () => HttpResponse.json(
			{ code: 'INVALID_TRANSITION', message: 'Договор уже подписан' },
			{ status: 409 },
		)));
		const { user } = renderWithProviders(<SignContract applicationId="a-1" contract={contractFixture} />);

		await user.click(screen.getByRole('button', { name: 'Подписать договор' }));
		await user.click(within(screen.getByRole('dialog')).getByRole('button', { name: 'Подтверждаю' }));

		expect(await within(screen.getByRole('dialog')).findByText('Договор уже подписан')).toBeInTheDocument();
	});
});
