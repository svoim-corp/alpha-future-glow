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
	limit: 1000000,
	prelimRate: 22,
	prelimPayment: 77800,
	amountFits: false,
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

	it('фиксирует согласие, запрашивает решение и показывает лимит и проверки', async () => {
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

		expect(await screen.findByText('Предварительный лимит')).toBeInTheDocument();
		expect(consent).toHaveBeenCalledWith(expect.objectContaining({ given: true }));
		expect(screen.getByText('Внимание: 1 арбитражное дело, некритично')).toBeInTheDocument();
		expect(screen.getByText('Запрошенная сумма выше лимита')).toBeInTheDocument();
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

	it('выбор «уменьшить сумму» отправляет лимит как новую сумму и сообщает мастеру о выборе', async () => {
		const choice = vi.fn();
		const onChosen = vi.fn();
		server.use(
			http.post(apiUrl('/applications/a-1031/prelim'), () => HttpResponse.json(limitBody)),
			http.post(apiUrl('/applications/a-1031/prelim/choice'), async ({ request }) => {
				choice(await request.json());

				return HttpResponse.json({ stage: 2, params: { amount: 1000000, term: 24 }, version: 9 });
			}),
		);
		const app = { ...fresh, consents: { bki: { given: true, at: '2026-10-20T10:05:00Z' } } };
		const { user } = renderWithProviders(<RequestPrelimDecision application={app} onChosen={onChosen} />);

		await user.click(screen.getByRole('button', { name: 'Получить предварительное решение' }));
		await user.click(await screen.findByRole('button', { name: 'Продолжить к полной заявке' }));

		await waitFor(() => expect(onChosen).toHaveBeenCalled());
		expect(choice).toHaveBeenCalledWith({ choice: 'REDUCE_AMOUNT', amount: 1000000 });
	});

	it('выбор «оставить сумму» отправляет KEEP_AMOUNT без суммы и тоже сообщает мастеру', async () => {
		const choice = vi.fn();
		const onChosen = vi.fn();
		server.use(
			http.post(apiUrl('/applications/a-1031/prelim'), () => HttpResponse.json(limitBody)),
			http.post(apiUrl('/applications/a-1031/prelim/choice'), async ({ request }) => {
				choice(await request.json());

				return HttpResponse.json({ stage: 2, params: { amount: 1500000, term: 24 }, version: 9 });
			}),
		);
		const app = { ...fresh, consents: { bki: { given: true, at: '2026-10-20T10:05:00Z' } } };
		const { user } = renderWithProviders(<RequestPrelimDecision application={app} onChosen={onChosen} />);

		await user.click(screen.getByRole('button', { name: 'Получить предварительное решение' }));
		await user.click(await screen.findByRole('radio', { name: /Оставить запрошенную сумму/ }));
		await user.click(screen.getByRole('button', { name: 'Продолжить к полной заявке' }));

		await waitFor(() => expect(choice).toHaveBeenCalledWith({ choice: 'KEEP_AMOUNT' }));
		expect(onChosen).toHaveBeenCalled();
	});

	it('если сумма в пределах лимита, выбор не предлагается', async () => {
		server.use(http.post(apiUrl('/applications/a-1031/prelim'), () => HttpResponse.json({
			...limitBody,
			limit: 3000000,
			amountFits: true,
		})));
		const app = { ...fresh, consents: { bki: { given: true, at: '2026-10-20T10:05:00Z' } } };
		const { user } = renderWithProviders(<RequestPrelimDecision application={app} />);

		await user.click(screen.getByRole('button', { name: 'Получить предварительное решение' }));

		expect(await screen.findByText('Предварительный лимит')).toBeInTheDocument();
		expect(screen.queryByRole('button', { name: 'Продолжить к полной заявке' })).not.toBeInTheDocument();
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
				limit: 1000000,
				prelimRate: 22,
				amountFits: false,
				choice: null,
			},
		};
		renderWithProviders(<RequestPrelimDecision application={app} />);

		expect(screen.getByText('Предварительный лимит')).toBeInTheDocument();
		expect(screen.getByRole('radio', { name: 'Уменьшить сумму до лимита' })).toBeInTheDocument();
	});
});
