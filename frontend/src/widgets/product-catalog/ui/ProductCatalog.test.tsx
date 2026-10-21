import {
	goalsFixture,
	matchFixture,
	productsFixture,
	workingDetails,
} from '@test/fixtures/products';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen, within } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { ProductCatalog } from './ProductCatalog';

const calculation = {
	rate: { min: 20, max: 28 },
	payment: { min: 76300, max: 82300 },
	gracePayment: null,
	withinProductLimits: true,
};

const useHappyPath = () => {
	server.use(
		http.get(apiUrl('/products/goals'), () => HttpResponse.json(goalsFixture)),
		http.get(apiUrl('/products'), () => HttpResponse.json(productsFixture)),
		http.get(apiUrl('/products/WORKING'), () => HttpResponse.json(workingDetails)),
		http.post(apiUrl('/products/match'), () => HttpResponse.json(matchFixture)),
		http.post(apiUrl('/calculator/payment'), () => HttpResponse.json(calculation)),
	);
};

describe('ProductCatalog', () => {
	it('подсвечивает подходящий продукт и показывает, что сумма и срок перейдут в заявку', async () => {
		useHappyPath();
		renderWithProviders(<ProductCatalog />);

		expect(await screen.findByText(/Подходят: Оборотный кредит/)).toBeInTheDocument();
		expect(screen.getByText('Подходит под задачу')).toBeInTheDocument();
		expect(screen.getAllByRole('article')).toHaveLength(3);
	});

	it('по клику на «Условия» открывает панель с условиями и кнопкой «Оформить»', async () => {
		useHappyPath();
		const { user } = renderWithProviders(<ProductCatalog />);

		await user.click(await screen.findByRole('button', { name: 'Условия: Оборотный кредит' }));

		const panel = await screen.findByRole('region', { name: 'Условия продукта' });
		expect(await within(panel).findByRole('heading', { name: 'Оборотный кредит' })).toBeInTheDocument();
		expect(within(panel).getByText('Что понадобится')).toBeInTheDocument();
		expect(within(panel).getByText('Согласие на запрос в бюро кредитных историй')).toBeInTheDocument();
		expect(within(panel).getByRole('button', { name: 'Оформить' })).toBeEnabled();
	});

	it('до выбора продукта панель подсказывает, что делать', async () => {
		useHappyPath();
		renderWithProviders(<ProductCatalog />);

		expect(await screen.findByText(/Выберите продукт в списке/)).toBeInTheDocument();
	});

	it('при неверной сумме показывает ошибку под полем и блокирует «Оформить»', async () => {
		useHappyPath();
		const { user } = renderWithProviders(<ProductCatalog />);

		await user.click(await screen.findByRole('button', { name: 'Условия: Оборотный кредит' }));
		const amount = screen.getByLabelText('Сколько нужно, ₽');
		await user.clear(amount);

		expect(await screen.findByText('Введите сумму в рублях целым числом')).toBeInTheDocument();
		const panel = screen.getByRole('region', { name: 'Условия продукта' });
		expect(await within(panel).findByRole('button', { name: 'Оформить' })).toBeDisabled();
	});

	it('овердрафт показывает лимит вместо расчёта по сроку', async () => {
		useHappyPath();
		server.use(http.get(apiUrl('/products/OVERDRAFT'), () => HttpResponse.json({
			...workingDetails,
			code: 'OVERDRAFT',
			name: 'Овердрафт',
			term: null,
			whatYouNeed: ['ИНН'],
		})));
		const { user } = renderWithProviders(<ProductCatalog />);

		await user.click(await screen.findByRole('button', { name: 'Условия: Овердрафт' }));

		expect(await screen.findByText('Лимит на 12 месяцев')).toBeInTheDocument();
	});

	it('при ошибке списка продуктов предлагает повторить', async () => {
		useHappyPath();
		server.use(http.get(apiUrl('/products'), () => HttpResponse.json(
			{ code: 'UPSTREAM_UNAVAILABLE', message: 'down' },
			{ status: 502 },
		), { once: true }));
		const { user } = renderWithProviders(<ProductCatalog />);

		expect(await screen.findByText('Не удалось загрузить продукты.')).toBeInTheDocument();
		await user.click(screen.getByRole('button', { name: 'Повторить' }));

		expect(await screen.findAllByRole('article')).toHaveLength(3);
	});

	it('если ничего не подошло, подсказывает изменить сумму или срок', async () => {
		useHappyPath();
		server.use(http.post(apiUrl('/products/match'), () => HttpResponse.json(
			matchFixture.map((item) => ({ ...item, fits: false })),
		)));
		renderWithProviders(<ProductCatalog />);

		expect(await screen.findByText(/продуктов не нашлось/)).toBeInTheDocument();
	});
});
