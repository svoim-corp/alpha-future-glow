import { makePlan } from '@test/fixtures/plans';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { chooseOption } from '@test/select';
import { screen, waitFor, within } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { CreatePlan } from './CreatePlan';
import { PlanActions } from './PlanActions';

beforeEach(() => {
	server.use(http.get(apiUrl('/products/goals'), () => HttpResponse.json([{ code: 'EQUIPMENT', name: 'Оборудование' }])));
});
const openCreate = async () => {
	const result = renderWithProviders(<CreatePlan />);
	await result.user.click(screen.getByRole('button', { name: 'Новый план' }));
	await waitFor(() => expect(screen.getByRole('button', { name: 'Сохранить план' })).toBeEnabled());

	return result;
};

describe('формы Planner с серверным контрактом', () => {
	it('не создаёт план без обязательной цели', async () => {
		const submit = vi.fn();
		server.use(http.post(apiUrl('/plans'), () => { submit(); return HttpResponse.json(makePlan()); }));
		const { user } = await openCreate();
		await user.type(screen.getByLabelText('Название'), 'Печь');
		await user.type(screen.getByLabelText('Ориентировочный бюджет, ₽'), '100000');
		await user.click(screen.getByRole('button', { name: 'Сохранить план' }));
		expect(await screen.findByText('Опишите цель', { selector: 'p' })).toBeInTheDocument();
		expect(submit).not.toHaveBeenCalled();
	});

	it('создаёт план со свободной целью независимо от справочника', async () => {
		const sent = vi.fn();
		server.use(
			http.get(apiUrl('/products/goals'), () => new HttpResponse(null, { status: 503 })),
			http.post(apiUrl('/plans'), async ({ request }) => {
				sent(await request.json());

				return HttpResponse.json(makePlan());
			}),
		);
		const { user } = await openCreate();
		await user.type(screen.getByLabelText('Название'), 'Пекарня');
		await user.type(screen.getByLabelText('Цель'), '  Открыть вторую пекарню  ');
		await user.type(screen.getByLabelText('Ориентировочный бюджет, ₽'), '1200000');
		expect(screen.getByLabelText('Ориентировочный бюджет, ₽')).toHaveValue('1 200 000');
		await user.click(screen.getByRole('button', { name: 'Сохранить план' }));
		await waitFor(() => expect(sent).toHaveBeenCalledWith(expect.objectContaining({
			goal: 'Открыть вторую пекарню', budgetRub: 1200000,
		})));
	});

	it('серверная ошибка поля сохраняет введённое и отображается рядом с полем', async () => {
		server.use(http.post(apiUrl('/plans'), () => HttpResponse.json({
			code: 'VALIDATION_ERROR', message: 'Проверьте бюджет',
			fields: [{ field: 'budgetRub', code: 'INVALID', message: 'Проверьте сумму' }],
		}, { status: 400 })));
		const { user } = await openCreate();
		await user.type(screen.getByLabelText('Название'), 'Печь');
		await user.type(screen.getByLabelText('Ориентировочный бюджет, ₽'), '100000');
		await user.type(screen.getByLabelText('Цель'), 'Открыть вторую пекарню');
		await user.click(screen.getByRole('button', { name: 'Сохранить план' }));
		expect(await screen.findByText('Проверьте сумму')).toBeInTheDocument();
		expect(screen.getByLabelText('Название')).toHaveValue('Печь');
		expect(screen.getByRole('dialog')).toBeInTheDocument();
	});

	it('редактирование очищает заметку и даты по серверному PATCH, сохраняя email-флаг', async () => {
		const plan = makePlan({ note: 'Старая заметка', notifications: { inApp: true, email: true } });
		const sent = vi.fn();
		server.use(http.patch(apiUrl(`/plans/${plan.id}`), async ({ request }) => {
			sent(await request.json());

			return HttpResponse.json({ ...plan, note: null, dateMode: 'UNSCHEDULED', startDate: null, endDate: null });
		}));
		const { user } = renderWithProviders(<PlanActions plan={plan} />);
		await user.click(screen.getByRole('button', { name: 'Редактировать план' }));
		await user.clear(screen.getByLabelText('Заметка'));
		await user.click(screen.getByRole('radio', { name: 'Без даты' }));
		await user.click(screen.getByRole('button', { name: 'Сохранить план' }));
		await waitFor(() => expect(sent).toHaveBeenCalledWith({ note: '', dateMode: 'UNSCHEDULED' }));
		await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
	});

	it('позволяет изменить существующий план без справочника, не меняя его цель', async () => {
		server.use(http.get(apiUrl('/products/goals'), () => new HttpResponse(null, { status: 404 })));
		const plan = makePlan();
		const sent = vi.fn();
		server.use(http.patch(apiUrl(`/plans/${plan.id}`), async ({ request }) => {
			sent(await request.json());

			return HttpResponse.json(plan);
		}));
		const { user } = renderWithProviders(<PlanActions plan={plan} />);
		await user.click(screen.getByRole('button', { name: 'Редактировать план' }));
		expect(screen.getByLabelText('Цель')).toHaveValue(plan.goal);
		await user.clear(screen.getByLabelText('Название'));
		await user.type(screen.getByLabelText('Название'), 'Новое название');
		await user.click(screen.getByRole('button', { name: 'Сохранить план' }));
		await waitFor(() => expect(sent).toHaveBeenCalledWith({ title: 'Новое название' }));
	});

	it('не отправляет PATCH без изменений и предупреждает при закрытии изменённой формы', async () => {
		const { user } = renderWithProviders(<PlanActions plan={makePlan()} />);
		await user.click(screen.getByRole('button', { name: 'Редактировать план' }));
		await user.click(screen.getByRole('button', { name: 'Сохранить план' }));
		expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
		await user.click(screen.getByRole('button', { name: 'Редактировать план' }));
		await user.type(screen.getByLabelText('Заметка'), 'Новая заметка');
		await user.click(within(screen.getByRole('dialog')).getByRole('button', { name: 'Отмена' }));
		expect(screen.getByText('Не сохранять изменения?')).toBeInTheDocument();
		await user.click(screen.getByRole('button', { name: 'Продолжить заполнение' }));
		expect(screen.getByLabelText('Заметка')).toHaveValue('Новая заметка');
		await user.click(screen.getByRole('button', { name: 'Закрыть' }));
		await user.click(screen.getByRole('button', { name: 'Не сохранять' }));
		expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
	});

	it('не закрывает форму во время запроса и не отправляет повторное сохранение', async () => {
		const sent = vi.fn();
		let respond: () => void = () => undefined;
		const pendingResponse = new Promise<void>((resolve) => { respond = resolve; });
		server.use(http.post(apiUrl('/plans'), async () => {
			sent();
			await pendingResponse;

			return HttpResponse.json(makePlan());
		}));
		const { user } = await openCreate();
		await user.type(screen.getByLabelText('Название'), 'Печь');
		await user.type(screen.getByLabelText('Ориентировочный бюджет, ₽'), '100000');
		await user.type(screen.getByLabelText('Цель'), 'Открыть вторую пекарню');
		await user.click(screen.getByRole('button', { name: 'Сохранить план' }));
		try {
			await waitFor(() => expect(sent).toHaveBeenCalledTimes(1));
			expect(screen.getByRole('button', { name: 'Сохранить план' })).toBeDisabled();
			await user.click(screen.getByRole('button', { name: 'Сохранить план' }));
			await user.click(screen.getByRole('button', { name: 'Закрыть' }));
			expect(screen.getByRole('dialog')).toBeInTheDocument();
		} finally {
			respond();
		}
		await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
		expect(sent).toHaveBeenCalledTimes(1);
	});

	it('не подменяет состояние успешным при серверной ошибке изменения', async () => {
		server.use(http.patch(apiUrl('/plans/:id'), () => HttpResponse.json({ code: 'NOT_FOUND', message: 'План удалён' }, { status: 404 })));
		const { user } = renderWithProviders(<PlanActions plan={makePlan()} />);
		await chooseOption(user, 'Состояние плана', 'В архиве');
		expect(await screen.findByText('План удалён')).toBeInTheDocument();
		expect(screen.getByLabelText('Состояние плана')).toHaveTextContent('Активен');
	});
});
