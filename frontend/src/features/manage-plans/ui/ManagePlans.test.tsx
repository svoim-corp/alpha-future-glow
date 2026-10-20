import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen, waitFor, within } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { CreatePlan } from './CreatePlan';
import { DeletePlan } from './DeletePlan';

const goals = http.get(apiUrl('/products/goals'), () => HttpResponse.json([{ code: 'EQUIPMENT', name: 'Оборудование' }]));

describe('CreatePlan', () => {
	it('идея без даты сохраняется без дат', async () => {
		const body = vi.fn();
		server.use(goals, http.post(apiUrl('/plans'), async ({ request }) => {
			body(await request.json());

			return HttpResponse.json({ id: 'p-9' }, { status: 201 });
		}));
		const { user } = renderWithProviders(<CreatePlan />);

		await user.click(screen.getByRole('button', { name: 'Новый план' }));
		const dialog = screen.getByRole('dialog');
		await user.type(within(dialog).getByLabelText('Название'), 'Новая печь');
		await user.type(within(dialog).getByLabelText('Ориентировочный бюджет, ₽'), '1200000');
		await user.click(within(dialog).getByRole('button', { name: 'Сохранить план' }));

		await waitFor(() => expect(body).toHaveBeenCalledWith({
			title: 'Новая печь',
			note: null,
			goal: null,
			budgetRub: 1200000,
			dateMode: 'UNSCHEDULED',
			startDate: null,
			endDate: null,
			notifications: { inApp: true, email: false },
		}));
		await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
	});

	it('период: показывает ошибку обратного диапазона дат', async () => {
		server.use(goals);
		const { user } = renderWithProviders(<CreatePlan />);

		await user.click(screen.getByRole('button', { name: 'Новый план' }));
		const dialog = screen.getByRole('dialog');
		await user.type(within(dialog).getByLabelText('Название'), 'Ремонт');
		await user.type(within(dialog).getByLabelText('Ориентировочный бюджет, ₽'), '100');
		await user.click(within(dialog).getByRole('radio', { name: 'Период' }));
		await user.type(within(dialog).getByLabelText('Начало'), '2027-06-10');
		await user.type(within(dialog).getByLabelText('Окончание'), '2027-06-01');
		await user.click(within(dialog).getByRole('button', { name: 'Сохранить план' }));

		expect(await within(dialog).findByText('Окончание не может быть раньше начала')).toBeInTheDocument();
	});

	it('один день отправляет одинаковые начало и окончание', async () => {
		const body = vi.fn();
		server.use(goals, http.post(apiUrl('/plans'), async ({ request }) => {
			body(await request.json());

			return HttpResponse.json({ id: 'p-9' }, { status: 201 });
		}));
		const { user } = renderWithProviders(<CreatePlan />);

		await user.click(screen.getByRole('button', { name: 'Новый план' }));
		const dialog = screen.getByRole('dialog');
		await user.type(within(dialog).getByLabelText('Название'), 'Закупка');
		await user.type(within(dialog).getByLabelText('Ориентировочный бюджет, ₽'), '50000');
		await user.click(within(dialog).getByRole('radio', { name: 'Один день' }));
		await user.type(within(dialog).getByLabelText('Дата'), '2026-11-05');
		await user.click(within(dialog).getByRole('button', { name: 'Сохранить план' }));

		await waitFor(() => expect(body).toHaveBeenCalledWith(expect.objectContaining({
			dateMode: 'DAY',
			startDate: '2026-11-05',
			endDate: '2026-11-05',
		})));
	});
});

describe('DeletePlan', () => {
	it('удаляет после подтверждения', async () => {
		const spy = vi.fn();
		server.use(http.delete(apiUrl('/plans/p-1'), () => {
			spy();

			return new HttpResponse(null, { status: 204 });
		}));
		const { user } = renderWithProviders(<DeletePlan id="p-1" title="Ремонт" />);

		await user.click(screen.getByRole('button', { name: 'Удалить план: Ремонт' }));
		await user.click(within(screen.getByRole('dialog')).getByRole('button', { name: 'Удалить' }));

		await waitFor(() => expect(spy).toHaveBeenCalledTimes(1));
	});
});
