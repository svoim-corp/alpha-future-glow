import { fullApplication } from '@test/fixtures/applications';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { RequestPrelimDecision } from './RequestPrelimDecision';

const fresh = { ...fullApplication, consents: {}, prelim: null, stage: 1 as const };
const limitBody = {
	outcome: 'LIMIT',
	requestedAmount: 1500000,
	limitNoStatement: 500000,
	limitWithStatement: 3000000,
	prelimRate: 22,
	prelimPayment: 77800,
	needsChoice: true,
	checks: [
		{ source: 'EGRUL', status: 'OK', label: 'Компания действует, 11 лет' },
		{ source: 'COURTS', status: 'WARN', label: '1 арбитражное дело, некритично' },
	],
};

describe('RequestPrelimDecision', () => {
	it('кнопка недоступна без согласия на БКИ', async () => {
		const { user } = renderWithProviders(<RequestPrelimDecision application={fresh} />);
		const button = screen.getByRole('button', { name: 'Получить предварительное решение' });

		expect(button).toBeDisabled();
		await user.click(screen.getByRole('checkbox'));
		expect(button).toBeEnabled();
	});

	it('без подтверждённых полномочий согласие заблокировано', () => {
		const app = { ...fresh, authority: { type: 'MCHD' as const, verified: false } };
		renderWithProviders(<RequestPrelimDecision application={app} />);

		expect(screen.getByRole('checkbox')).toBeDisabled();
		expect(screen.getByText('Сначала подтвердите полномочия')).toBeInTheDocument();
	});

	it('фиксирует согласие, запрашивает решение и показывает два лимита и проверки', async () => {
		const consent = vi.fn();
		server.use(
			http.post(apiUrl('/applications/a-1031/consents/bki'), async ({ request }) => {
				consent(await request.json());

				return HttpResponse.json({ given: true, at: '2026-10-20T10:05:00Z' }, { status: 201 });
			}),
			http.post(apiUrl('/applications/a-1031/prelim'), () => HttpResponse.json(limitBody)),
		);
		const { user } = renderWithProviders(<RequestPrelimDecision application={fresh} />);

		await user.click(screen.getByRole('checkbox'));
		await user.click(screen.getByRole('button', { name: 'Получить предварительное решение' }));

		expect(await screen.findByText('Лимит без выписки')).toBeInTheDocument();
		expect(consent).toHaveBeenCalledWith(expect.objectContaining({ given: true }));
		expect(screen.getByText('Внимание: 1 арбитражное дело, некритично')).toBeInTheDocument();
		expect(screen.getByText('Запрошенная сумма выше лимита без выписки')).toBeInTheDocument();
	});

	it('стоп-отказ объясняет причину и что делать', async () => {
		server.use(http.post(apiUrl('/applications/a-1031/prelim'), () => HttpResponse.json({
			outcome: 'STOP',
			stopReason: 'BUSINESS_TOO_YOUNG',
			stopText: 'Бизнес зарегистрирован меньше 6 месяцев назад',
			whatToDo: ['Вернуться, когда бизнесу будет 6 месяцев'],
		})));
		const app = { ...fresh, consents: { bki: { given: true, at: '2026-10-20T10:05:00Z' } } };
		const { user } = renderWithProviders(<RequestPrelimDecision application={app} />);

		await user.click(screen.getByRole('button', { name: 'Получить предварительное решение' }));

		expect(await screen.findByText('Бизнес зарегистрирован меньше 6 месяцев назад')).toBeInTheDocument();
		expect(screen.getByText('Вернуться, когда бизнесу будет 6 месяцев')).toBeInTheDocument();
	});

	it('выбор «уменьшить сумму» отправляет новую сумму и переходит к этапу 2', async () => {
		const choice = vi.fn();
		const onStage2 = vi.fn();
		server.use(
			http.post(apiUrl('/applications/a-1031/prelim'), () => HttpResponse.json(limitBody)),
			http.post(apiUrl('/applications/a-1031/prelim/choice'), async ({ request }) => {
				choice(await request.json());

				return HttpResponse.json({
					stage: 2,
					statementRequired: false,
					params: { amount: 500000, term: 24 },
					version: 9,
				});
			}),
		);
		const app = { ...fresh, consents: { bki: { given: true, at: '2026-10-20T10:05:00Z' } } };
		const { user } = renderWithProviders(<RequestPrelimDecision application={app} onStage2={onStage2} />);

		await user.click(screen.getByRole('button', { name: 'Получить предварительное решение' }));
		await user.click(await screen.findByRole('radio', { name: 'Уменьшить сумму до лимита без выписки' }));
		await user.type(screen.getByLabelText('Новая сумма, ₽'), '500000');
		await user.click(screen.getByRole('button', { name: 'Продолжить к полной заявке' }));

		await waitFor(() => expect(onStage2).toHaveBeenCalled());
		expect(choice).toHaveBeenCalledWith({ choice: 'REDUCE_AMOUNT', amount: 500000 });
	});

	it('показывает ошибку, если сервис проверок недоступен', async () => {
		server.use(http.post(apiUrl('/applications/a-1031/prelim'), () => HttpResponse.json(
			{ code: 'UPSTREAM_UNAVAILABLE', message: 'Бюро кредитных историй недоступно' },
			{ status: 502 },
		)));
		const app = { ...fresh, consents: { bki: { given: true, at: '2026-10-20T10:05:00Z' } } };
		const { user } = renderWithProviders(<RequestPrelimDecision application={app} />);

		await user.click(screen.getByRole('button', { name: 'Получить предварительное решение' }));

		expect(await screen.findByText('Бюро кредитных историй недоступно')).toBeInTheDocument();
	});

	it('по сохранённому лимиту без выбора показывает варианты после перезагрузки', () => {
		const app = {
			...fresh,
			prelim: {
				outcome: 'LIMIT' as const,
				limitNoStatement: 500000,
				limitWithStatement: 3000000,
				prelimRate: 22,
				choice: null,
			},
		};
		renderWithProviders(<RequestPrelimDecision application={app} />);

		expect(screen.getByText('Лимит с выпиской')).toBeInTheDocument();
		expect(screen.getByRole('radio', { name: 'Подтвердить обороты выпиской по счёту' })).toBeInTheDocument();
	});
});
