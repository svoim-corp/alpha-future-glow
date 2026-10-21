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
	it('показывает все продукты, а цель-чип оставляет подходящие и подсвечивает их', async () => {
		useHappyPath();
		const { user } = renderWithProviders(<ProductCatalog />);

		expect(await screen.findAllByRole('article')).toHaveLength(3);
		expect(screen.getByRole('button', { name: 'Все предложения', pressed: true })).toBeInTheDocument();

		await user.click(screen.getByRole('button', { name: 'Пополнить оборотные средства' }));

		expect(await screen.findByText(/Подходят: Оборотный кредит/)).toBeInTheDocument();
		expect(screen.getAllByRole('article')).toHaveLength(2);
		expect(screen.getByText('Подходит под задачу')).toBeInTheDocument();
	});

	it('по клику на «Посмотреть условия» открывает диалог с условиями и кнопкой «Оформить»', async () => {
		useHappyPath();
		const { user } = renderWithProviders(<ProductCatalog />);

		await user.click(await screen.findByRole('button', { name: 'Условия: Оборотный кредит' }));

		const dialog = await screen.findByRole('dialog');
		expect(await within(dialog).findByText('Что понадобится')).toBeInTheDocument();
		expect(within(dialog).getByText('Согласие на запрос в бюро кредитных историй')).toBeInTheDocument();
		expect(within(dialog).getByRole('button', { name: 'Оформить' })).toBeEnabled();
	});

	it('калькулятор считает платёж с сервера и меняет продукт', async () => {
		useHappyPath();
		const { user } = renderWithProviders(<ProductCatalog />);

		expect(await screen.findByText(/76\s300\s–\s82\s300/)).toBeInTheDocument();
		expect(screen.getByLabelText('Срок')).toBeInTheDocument();

		await user.selectOptions(screen.getByLabelText('Продукт'), 'OVERDRAFT');

		expect(await screen.findByText('Проценты только за дни в минусе')).toBeInTheDocument();
		expect(screen.queryByLabelText('Срок')).not.toBeInTheDocument();
	});

	it('овердрафт в диалоге показывает лимит вместо расчёта по сроку', async () => {
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

		expect(await screen.findByText('Лимит на 12 месяцев', { selector: 'p' })).toBeInTheDocument();
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
		const { user } = renderWithProviders(<ProductCatalog />);

		await user.click(await screen.findByRole('button', { name: 'Закрыть кассовый разрыв' }));

		expect(await screen.findByText(/продуктов не нашлось/)).toBeInTheDocument();
	});
});
