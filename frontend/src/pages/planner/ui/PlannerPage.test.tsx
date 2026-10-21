import { type BusinessPlan } from '@entities/plan';
import { makePlan } from '@test/fixtures/plans';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { chooseOption } from '@test/select';
import { screen, waitFor, within } from '@testing-library/react';
import { delay, http, HttpResponse } from 'msw';

import { PlannerPage } from './PlannerPage';

const list = (items: BusinessPlan[]) => ({
	items,
	summary: {
		total: items.length,
		withPeriod: items.filter((p) => p.dateMode !== 'UNSCHEDULED').length,
		unscheduled: items.filter((p) => p.dateMode === 'UNSCHEDULED').length,
		budgetRub: items.reduce((sum, p) => sum + p.budgetRub, 0),
	},
});

let plans: BusinessPlan[];
beforeEach(() => {
	vi.useFakeTimers({ toFake: ['Date'] });
	vi.setSystemTime(new Date('2026-10-21T12:00:00Z'));
	plans = [makePlan(), makePlan({ id: 'p-other', title: 'Ремонт', state: 'PAUSED' }), makePlan({
		id: 'p-idea', title: 'Идея без срока', dateMode: 'UNSCHEDULED', startDate: null, endDate: null,
	})];
	server.use(
		http.get(apiUrl('/products/goals'), () => HttpResponse.json([{ code: 'EQUIPMENT', name: 'Оборудование' }])),
		http.get(apiUrl('/plans'), ({ request }) => {
			const state = new URL(request.url).searchParams.get('state');

			return HttpResponse.json(list(plans.filter((p) => state ? p.state === state : p.state !== 'ARCHIVED')));
		}),
		http.patch(apiUrl('/plans/:id'), async ({ request, params }) => {
			const body = await request.json() as Partial<BusinessPlan>;
			plans = plans.map((p) => p.id === params.id ? { ...p, ...body } : p);

			return HttpResponse.json(plans.find((p) => p.id === params.id));
		}),
		http.delete(apiUrl('/plans/:id'), ({ params }) => {
			plans = plans.filter((p) => p.id !== params.id);

			return new HttpResponse(null, { status: 204 });
		}),
	);
});
afterEach(() => vi.useRealTimers());

const day = () => screen.getByRole('button', { name: /21 октября 2026.*планов: 2/i });

describe('новый календарь планера', () => {
	it('открывает текущий месяц, а не месяц первого плана; считает сводку по ответу сервера', async () => {
		plans = [makePlan({ startDate: '2027-06-01', endDate: '2027-08-31' })];
		renderWithProviders(<PlannerPage />);

		expect(await screen.findByRole('heading', { name: 'Октябрь 2026' })).toBeInTheDocument();
		expect(screen.getByText('Общий ориентир бюджета')).toBeInTheDocument();
		expect(screen.queryByRole('region', { name: 'Детали плана' })).not.toBeInTheDocument();
	});

	it('день показывает все пересечения, детали открываются только при выборе плана', async () => {
		const { user } = renderWithProviders(<PlannerPage />);
		await screen.findByRole('heading', { name: 'Октябрь 2026' });
		await user.click(day());
		const region = screen.getByRole('region', { name: /Планы на 21 октября 2026/ });
		expect(within(region).getAllByRole('button')).toHaveLength(2);
		await user.click(within(region).getByRole('button', { name: /Открыть план: Ремонт/ }));
		expect(screen.getByRole('dialog', { name: 'Ремонт' })).toBeInTheDocument();
		expect(within(screen.getByRole('region', { name: 'Детали плана' }))
			.getByRole('combobox', { name: 'Состояние плана' })).toHaveTextContent('На паузе');
	});

	it('переключает месяцы и возвращает на сегодняшний московский день', async () => {
		const { user } = renderWithProviders(<PlannerPage />);
		await screen.findByRole('heading', { name: 'Октябрь 2026' });
		await user.click(screen.getByRole('button', { name: 'Следующий месяц' }));
		expect(screen.getByRole('heading', { name: 'Ноябрь 2026' })).toBeInTheDocument();
		await user.click(screen.getByRole('button', { name: 'Сегодня' }));
		expect(day()).toHaveAttribute('aria-current', 'date');
		expect(day()).toHaveAttribute('aria-pressed', 'true');
	});

	it('выбор месяца и года перестраивает календарь и закрывает старые детали', async () => {
		const { user } = renderWithProviders(<PlannerPage />);
		await screen.findByRole('heading', { name: 'Октябрь 2026' });
		await user.click(day());
		await user.click(screen.getByRole('button', { name: /Открыть план: Ремонт/ }));
		await user.keyboard('{Escape}');
		await chooseOption(user, 'Месяц', 'февраль');
		await chooseOption(user, 'Год', '2028');
		expect(screen.getByRole('heading', { name: 'Февраль 2028' })).toBeInTheDocument();
		expect(screen.getByRole('button', { name: /29 февраля 2028.*планов: 0/ })).toBeInTheDocument();
		expect(screen.queryByRole('region', { name: 'Детали плана' })).not.toBeInTheDocument();
		expect(screen.queryByRole('button', { name: 'Создать новый план' })).not.toBeInTheDocument();
	});

	it('пустой день предлагает создание с выбранной датой и сохраняет её в API', async () => {
		const submit = vi.fn();
		server.use(http.post(apiUrl('/plans'), async ({ request }) => {
			submit(await request.json());

			return HttpResponse.json(makePlan(), { status: 201 });
		}));
		const { user } = renderWithProviders(<PlannerPage />);
		await screen.findByRole('heading', { name: 'Октябрь 2026' });
		expect(screen.queryByRole('button', { name: 'Создать новый план' })).not.toBeInTheDocument();
		await user.click(screen.getByRole('button', { name: /25 октября 2026.*планов: 0/ }));
		await user.click(screen.getByRole('button', { name: 'Создать новый план' }));
		expect(screen.getByLabelText('Дата')).toHaveValue('25.10.2026');
		expect(screen.getByRole('radio', { name: 'Один день' })).toBeChecked();
		await user.type(screen.getByLabelText('Название'), 'Новый станок');
		await user.type(screen.getByLabelText('Ориентировочный бюджет, ₽'), '500000');
		await user.type(screen.getByLabelText('Цель'), 'Открыть вторую пекарню');
		await user.click(screen.getByRole('button', { name: 'Сохранить план' }));
		await waitFor(() => expect(submit).toHaveBeenCalledWith(expect.objectContaining({
			dateMode: 'DAY', startDate: '2026-10-25', endDate: '2026-10-25',
		})));
	});

	it('помещает сводку в начало и поясняет планер через подсказку заголовка', async () => {
		const { user } = renderWithProviders(<PlannerPage />);
		await screen.findByRole('heading', { name: 'Октябрь 2026' });
		expect(screen.queryByText('У каждого плана — своё время')).not.toBeInTheDocument();
		await user.hover(screen.getByRole('button', { name: 'Как пользоваться планером' }));
		expect(screen.getByRole('tooltip')).toHaveTextContent('Планер помогает сохранять идеи развития бизнеса');
	});

	it('идеи без даты доступны отдельно и не попадают в выбранный день', async () => {
		const { user } = renderWithProviders(<PlannerPage />);
		await screen.findByRole('heading', { name: 'Октябрь 2026' });
		await user.click(screen.getByRole('tab', { name: /Идеи без даты/ }));
		expect(screen.getByRole('article', { name: 'Идея без срока' })).toBeInTheDocument();
		expect(screen.queryByRole('article', { name: 'Ремонт' })).not.toBeInTheDocument();
	});

	it('архив запрашивается с фильтром сервера', async () => {
		plans.push(makePlan({ id: 'p-archive', title: 'Архивный план', state: 'ARCHIVED' }));
		const { user } = renderWithProviders(<PlannerPage />);
		await screen.findByRole('heading', { name: 'Октябрь 2026' });
		await user.click(screen.getByRole('tab', { name: /Архив/ }));
		expect(await screen.findByRole('article', { name: 'Архивный план' })).toBeInTheDocument();
		expect(screen.queryByRole('article', { name: 'Ремонт' })).not.toBeInTheDocument();
	});

	it('после смены состояния обновляет детали, а после удаления не оставляет старую карточку', async () => {
		const { user } = renderWithProviders(<PlannerPage />);
		await screen.findByRole('heading', { name: 'Октябрь 2026' });
		await user.click(day());
		await user.click(screen.getByRole('button', { name: /Открыть план: Закупка оборудования/ }));
		await chooseOption(user, 'Состояние плана', 'Завершён');
		await waitFor(() => expect(screen.getByLabelText('Состояние плана')).toHaveTextContent('Завершён'));
		await user.click(screen.getByRole('button', { name: 'Удалить план: Закупка оборудования' }));
		await user.click(within(screen.getByRole('dialog')).getByRole('button', { name: 'Удалить' }));
		await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
	});

	it('открывает всю карточку клавиатурой и возвращает фокус после закрытия', async () => {
		const { user } = renderWithProviders(<PlannerPage />);
		await screen.findByRole('heading', { name: 'Октябрь 2026' });
		expect(screen.queryByText('Всё начинается с плана')).not.toBeInTheDocument();
		expect(screen.queryByText('СНАЧАЛА ИДЕЯ. ЗАТЕМ — ВОЗМОЖНОСТИ.')).not.toBeInTheDocument();
		await user.click(screen.getByRole('tab', { name: /Все планы/ }));
		const card = within(screen.getByRole('article', { name: 'Ремонт' }));
		expect(card.queryByText('Подробнее')).not.toBeInTheDocument();
		const opener = card.getByRole('button', { name: 'Открыть план: Ремонт' });
		await user.click(opener);
		expect(screen.getByRole('dialog', { name: 'Ремонт' })).toBeInTheDocument();
		await user.keyboard('{Escape}');
		expect(opener).toHaveFocus();
		await user.keyboard('{Enter}');
		expect(screen.getByRole('dialog', { name: 'Ремонт' })).toBeInTheDocument();
		await user.keyboard('{Escape}');
		await user.keyboard(' ');
		expect(screen.getByRole('dialog', { name: 'Ремонт' })).toBeInTheDocument();
	});

	it('просмотр, редактирование и подтверждение удаления показывают по одному диалогу', async () => {
		const { user } = renderWithProviders(<PlannerPage />);
		await screen.findByRole('heading', { name: 'Октябрь 2026' });
		await user.click(screen.getByRole('button', { name: 'Открыть план: Закупка оборудования' }));
		await user.click(screen.getByRole('button', { name: 'Редактировать план' }));
		expect(screen.getAllByRole('dialog')).toHaveLength(1);
		expect(screen.getByRole('dialog', { name: 'Редактировать план' })).toBeInTheDocument();
		await user.click(screen.getByRole('button', { name: 'Отмена' }));
		expect(screen.getByRole('dialog', { name: 'Закупка оборудования' })).toBeInTheDocument();
		await user.click(screen.getByRole('button', { name: 'Удалить план: Закупка оборудования' }));
		expect(screen.getAllByRole('dialog')).toHaveLength(1);
		expect(screen.getByText(/Удаление безвозвратное/)).toBeInTheDocument();
		await user.click(screen.getByRole('button', { name: 'Отмена' }));
		expect(screen.getByRole('dialog', { name: 'Закупка оборудования' })).toBeInTheDocument();
	});

	it('возвращает план из архива через состояние сервера и не смешивает архив с рабочим списком', async () => {
		plans.push(makePlan({ id: 'p-archive', title: 'Архивный план', state: 'ARCHIVED' }));
		const { user } = renderWithProviders(<PlannerPage />);
		await screen.findByRole('heading', { name: 'Октябрь 2026' });
		await user.click(screen.getByRole('tab', { name: /Архив/ }));
		await user.click(await screen.findByRole('button', { name: 'Открыть план: Архивный план' }));
		await chooseOption(user, 'Состояние плана', 'Активен');
		expect(await screen.findByText('В архиве пока нет планов')).toBeInTheDocument();
		expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
		await user.click(screen.getByRole('tab', { name: /Все планы/ }));
		expect(await screen.findByRole('article', { name: 'Архивный план' })).toBeInTheDocument();
	});

	it('пустой набор оставляет рабочий календарь и кнопку создания', async () => {
		plans = [];
		renderWithProviders(<PlannerPage />);
		expect(await screen.findByRole('heading', { name: 'Октябрь 2026' })).toBeInTheDocument();
		expect(screen.getByText('Планов пока нет')).toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'Новый план' })).toBeInTheDocument();
	});

	it('ошибка справочника не скрывает планы', async () => {
		server.use(http.get(apiUrl('/products/goals'), () => new HttpResponse(null, { status: 503 })));
		renderWithProviders(<PlannerPage />);
		expect(await screen.findByRole('heading', { name: 'Октябрь 2026' })).toBeInTheDocument();
		expect(day()).toBeInTheDocument();
	});

	it('показывает загрузку и ошибку с повтором без подстановки старых данных', async () => {
		server.use(http.get(apiUrl('/plans'), async () => {
			await delay(50);

			return HttpResponse.json({ code: 'INTERNAL', message: 'Сбой планера' }, { status: 500 });
		}, { once: true }));
		const { user } = renderWithProviders(<PlannerPage />);
		expect(screen.getByRole('status')).toBeInTheDocument();
		expect(await screen.findByRole('alert')).toBeInTheDocument();
		await user.click(screen.getByRole('button', { name: 'Повторить' }));
		expect(await screen.findByRole('heading', { name: 'Октябрь 2026' })).toBeInTheDocument();
	});
});
