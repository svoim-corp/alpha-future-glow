import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen, waitFor, within } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { MakeDecision } from './MakeDecision';

const requested = { productCode: 'WORKING', amount: 1500000, term: 24 };
const ok = (spy: (body: unknown, key: string | null) => void) => http.post(
	apiUrl('/bank/applications/a-1/decision'),
	async ({ request }) => {
		spy(await request.json(), request.headers.get('Idempotency-Key'));

		return HttpResponse.json({ status: 'COUNTER_OFFER', decisionId: 'dec-1', mode: 'MANUAL' });
	},
);

describe('MakeDecision', () => {
	it('«Одобрить» отправляет запрошенные параметры с ключом идемпотентности', async () => {
		const spy = vi.fn();
		server.use(ok(spy));
		const { user } = renderWithProviders(<MakeDecision applicationId="a-1" requested={requested} defaultRate={20} />);

		await user.click(screen.getByRole('button', { name: 'Одобрить как запрошено' }));

		await waitFor(() => expect(spy).toHaveBeenCalledWith(
			{ type: 'APPROVE', ...requested, rate: 20 },
			expect.stringMatching(/^[0-9a-f-]{36}$/u),
		));
		expect(await screen.findByText('Решение сохранено')).toBeInTheDocument();
	});

	it('без рассчитанной ставки одобрить как запрошено нельзя: бэкенд требует ставку', () => {
		renderWithProviders(<MakeDecision applicationId="a-1" requested={requested} />);

		expect(screen.getByRole('button', { name: 'Одобрить как запрошено' })).toBeDisabled();
		expect(screen.getByText('Ставка не рассчитана')).toBeInTheDocument();
	});

	it('иные условия: ставка предзаполнена рекомендованной и обязательна', async () => {
		const spy = vi.fn();
		server.use(ok(spy));
		const { user } = renderWithProviders(<MakeDecision applicationId="a-1" requested={requested} />);

		await user.click(screen.getByRole('button', { name: 'Иные условия' }));
		const dialog = screen.getByRole('dialog');
		await user.type(within(dialog).getByLabelText('Объяснение для клиента'), 'Сумма снижена');
		await user.click(within(dialog).getByRole('button', { name: 'Предложить клиенту' }));

		expect(await within(dialog).findByText('Укажите ставку, % годовых')).toBeInTheDocument();
		expect(spy).not.toHaveBeenCalled();
	});

	it('иные условия: сумма и срок предзаполнены, объяснение обязательно', async () => {
		const spy = vi.fn();
		server.use(ok(spy));
		const { user } = renderWithProviders(<MakeDecision applicationId="a-1" requested={requested} />);

		await user.click(screen.getByRole('button', { name: 'Иные условия' }));
		const dialog = screen.getByRole('dialog');
		expect(within(dialog).getByLabelText('Сумма, ₽')).toHaveValue('1500000');
		await user.click(within(dialog).getByRole('button', { name: 'Предложить клиенту' }));
		expect(await within(dialog).findByText('Укажите сумму, срок и объяснение для клиента')).toBeInTheDocument();
		expect(spy).not.toHaveBeenCalled();

		await user.clear(within(dialog).getByLabelText('Сумма, ₽'));
		await user.type(within(dialog).getByLabelText('Сумма, ₽'), '1200000');
		await user.type(within(dialog).getByLabelText('Ставка, % годовых'), '21,5');
		await user.type(within(dialog).getByLabelText('Объяснение для клиента'), 'Сумма снижена');
		await user.click(within(dialog).getByRole('button', { name: 'Предложить клиенту' }));

		await waitFor(() => expect(spy).toHaveBeenCalledWith(
			{
				type: 'COUNTER',
				amount: 1200000,
				term: 24,
				productCode: 'WORKING',
				rate: 21.5,
				reason: 'Сумма снижена',
			},
			expect.any(String),
		));
	});

	it('отказ требует причину', async () => {
		const spy = vi.fn();
		server.use(ok(spy));
		const { user } = renderWithProviders(<MakeDecision applicationId="a-1" requested={requested} />);

		await user.click(screen.getByRole('button', { name: 'Отказать' }));
		const dialog = screen.getByRole('dialog');
		await user.click(within(dialog).getByRole('button', { name: 'Отказать' }));

		expect(await within(dialog).findByText('Укажите причину отказа — клиент её увидит')).toBeInTheDocument();
		expect(spy).not.toHaveBeenCalled();
	});

	it('условия вне рамок продукта: показывает ответ сервера в окне', async () => {
		server.use(http.post(apiUrl('/bank/applications/a-1/decision'), () => HttpResponse.json(
			{ code: 'BUSINESS_RULE', message: 'Сумма вне рамок продукта' },
			{ status: 422 },
		)));
		const { user } = renderWithProviders(<MakeDecision applicationId="a-1" requested={requested} defaultRate={20} />);

		await user.click(screen.getByRole('button', { name: 'Иные условия' }));
		const dialog = screen.getByRole('dialog');
		await user.type(within(dialog).getByLabelText('Объяснение для клиента'), 'Причина');
		await user.click(within(dialog).getByRole('button', { name: 'Предложить клиенту' }));

		expect(await within(dialog).findByText('Сумма вне рамок продукта')).toBeInTheDocument();
	});
});
