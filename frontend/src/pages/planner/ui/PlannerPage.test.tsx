import { plansFixture } from '@test/fixtures/plans';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen, within } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { PlannerPage } from './PlannerPage';

const base = () => server.use(
	http.get(apiUrl('/plans'), () => HttpResponse.json(plansFixture)),
	http.get(apiUrl('/products/goals'), () => HttpResponse.json([{ code: 'EQUIPMENT', name: 'Оборудование' }])),
);

describe('PlannerPage', () => {
	it('показывает метрики и выбранный план в календаре на месяце его начала', async () => {
		base();
		renderWithProviders(<PlannerPage />);

		expect(await screen.findByRole('heading', { name: /июнь 2027/ })).toBeInTheDocument();
		expect(within(screen.getByRole('region', { name: 'Выбранный план' })).getByText('Ремонт помещения'))
			.toBeInTheDocument();
		expect(screen.getByText('Общий ориентир бюджета')).toBeInTheDocument();
		expect(screen.getByText(/4\s250\s000/)).toBeInTheDocument();
	});

	it('дни периода плана подписаны для скринридера', async () => {
		base();
		renderWithProviders(<PlannerPage />);

		await screen.findByRole('heading', { name: /июнь 2027/ });

		expect(screen.getAllByText(/— Ремонт помещения/)).toHaveLength(30);
	});

	it('выбор другого плана переносит календарь на его месяц', async () => {
		base();
		const { user } = renderWithProviders(<PlannerPage />);

		await screen.findByRole('heading', { name: /июнь 2027/ });
		await user.click(screen.getByRole('button', { name: 'Запасы к зимнему сезону' }));

		expect(screen.getByRole('heading', { name: /ноябрь 2026/ })).toBeInTheDocument();
	});

	it('вкладка идей показывает только планы без даты и подбор', async () => {
		base();
		const { user } = renderWithProviders(<PlannerPage />);

		await user.click(await screen.findByRole('tab', { name: /Идеи без даты/ }));

		expect(screen.getByRole('article', { name: 'Новая печь для пекарни' })).toBeInTheDocument();
		expect(screen.queryByRole('article', { name: 'Ремонт помещения' })).not.toBeInTheDocument();
		expect(screen.getByText(/Подходит: Инвестиционный кредит/)).toBeInTheDocument();
	});

	it('без планов объясняет, как начать', async () => {
		server.use(http.get(apiUrl('/plans'), () => HttpResponse.json({
			items: [],
			summary: { total: 0, withPeriod: 0, unscheduled: 0, budgetRub: 0 },
		})), http.get(apiUrl('/products/goals'), () => HttpResponse.json([])));
		renderWithProviders(<PlannerPage />);

		expect(await screen.findByText('Планов пока нет')).toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'Новый план' })).toBeInTheDocument();
	});
});
