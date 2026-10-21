import { plansFixture } from '@test/fixtures/plans';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { PlanCard } from './PlanCard';
import { PlanRecommendations } from './PlanRecommendations';

const [repair, , oven] = plansFixture.items;

describe('PlanCard', () => {
	it('план с периодом показывает даты и бюджет', () => {
		renderWithProviders(<PlanCard plan={repair!} goalName="Недвижимость" />);

		expect(screen.getByText('План с периодом')).toBeInTheDocument();
		expect(screen.getByText(/01\.06\.2027 — 31\.08\.2027 · 2\s400\s000/)).toBeInTheDocument();
		expect(screen.getByText('Цель: Недвижимость')).toBeInTheDocument();
	});

	it('идея без даты помечена и показывает подбор', () => {
		renderWithProviders(<PlanCard plan={oven!} />);

		expect(screen.getByText('Идея без даты')).toBeInTheDocument();
		expect(screen.getByText(/Без даты · идея для развития/)).toBeInTheDocument();
		expect(screen.getByText(/Подходит: Инвестиционный кредит/)).toBeInTheDocument();
	});
});

describe('PlanRecommendations', () => {
	it('без подбора виджет не показывается', async () => {
		server.use(http.get(apiUrl('/plans/recommendations'), () => HttpResponse.json([])));
		const { container } = renderWithProviders(<PlanRecommendations />);

		await new Promise((resolve) => { setTimeout(resolve, 30); });

		expect(container).toBeEmptyDOMElement();
	});

	it('с подбором — объясняет и ведёт в планер', async () => {
		server.use(http.get(apiUrl('/plans/recommendations'), () => HttpResponse.json([{
			planId: 'p-3',
			planTitle: 'Новая печь для пекарни',
			budgetRub: 1200000,
			productCode: 'INVEST',
			productName: 'Инвестиционный кредит',
			reason: 'сумма и цель укладываются в рамки продукта',
		}])));
		renderWithProviders(<PlanRecommendations />);

		expect(await screen.findByText('Идеи становятся ближе')).toBeInTheDocument();
		expect(screen.getByRole('link', { name: /Все планы/ })).toHaveAttribute('href', '/planner');
		expect(screen.getByRole('link', { name: /Изучить/ })).toHaveAttribute('href', '/products');
		expect(screen.getByText(/не решение банка/)).toBeInTheDocument();
	});
});
