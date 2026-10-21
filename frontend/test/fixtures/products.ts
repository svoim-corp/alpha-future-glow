import { type Goal, type ProductDetails, type ProductMatch, type ProductSummary } from '@entities/product';

export const goalsFixture: Goal[] = [
	{ code: 'WORKING_CAPITAL', name: 'Пополнить оборотные средства' },
	{ code: 'EQUIPMENT', name: 'Купить оборудование' },
	{ code: 'CASH_GAP', name: 'Закрыть кассовый разрыв' },
];

export const productsFixture: ProductSummary[] = [
	{
		code: 'EXPRESS',
		name: 'Экспресс-кредит',
		shortText: 'Небольшая сумма на любые цели бизнеса',
		goals: ['WORKING_CAPITAL', 'EQUIPMENT'],
		amount: { min: 100000, max: 1000000 },
		term: { min: 3, max: 24 },
		rate: { min: 24, max: 30 },
	},
	{
		code: 'WORKING',
		name: 'Оборотный кредит',
		shortText: 'Закупки, зарплата, налоги, аренда',
		goals: ['WORKING_CAPITAL'],
		amount: { min: 300000, max: 10000000 },
		term: { min: 3, max: 36 },
		rate: { min: 20, max: 28 },
	},
	{
		code: 'OVERDRAFT',
		name: 'Овердрафт',
		shortText: 'Кассовые разрывы по расчётному счёту',
		goals: ['CASH_GAP'],
		amount: { min: 100000, max: 5000000 },
		term: null,
		rate: { min: 22, max: 30 },
	},
];

export const workingDetails: ProductDetails = {
	code: 'WORKING',
	name: 'Оборотный кредит',
	description: 'Разовая выдача на текущие нужды. Лимит считается от оборотов.',
	goals: ['WORKING_CAPITAL'],
	amount: { min: 300000, max: 10000000 },
	term: { min: 3, max: 36 },
	rate: { min: 20, max: 28 },
	requirements: { minBusinessAgeMonths: null, guarantorForLlc: false },
	whatYouNeed: ['Согласие на запрос в бюро кредитных историй'],
};

export const investDetails: ProductDetails = {
	code: 'INVEST',
	name: 'Инвестиционный кредит',
	description: 'Оборудование, помещение, расширение бизнеса',
	goals: ['EQUIPMENT', 'REAL_ESTATE', 'EXPANSION'],
	amount: { min: 1000000, max: 30000000 },
	term: { min: 12, max: 84 },
	rate: { min: 18, max: 26 },
	graceMaxMonths: 6,
	requirements: { minBusinessAgeMonths: 12, guarantorForLlc: true },
	limitRule: 'До 1,5 среднемесячных поступлений на счета в банке',
	whatYouNeed: ['Согласие на запрос в БКИ', 'Для ООО — поручитель из руководителя или учредителей и его согласие на запрос в БКИ'],
};

export const overdraftDetails: ProductDetails = {
	code: 'OVERDRAFT',
	name: 'Овердрафт',
	description: 'Минус по расчётному счёту, гасится поступлениями.',
	goals: ['CASH_GAP'],
	amount: { min: 100000, max: 5000000 },
	term: null,
	rate: { min: 22, max: 30 },
	requirements: { minBusinessAgeMonths: null, guarantorForLlc: false },
	whatYouNeed: ['Согласие на запрос в БКИ'],
};

export const matchFixture: ProductMatch[] = [
	{ code: 'EXPRESS', fits: false, reason: 'Сумма больше 1 000 000 ₽' },
	{ code: 'WORKING', fits: true, reason: 'Подходит по цели, сумме и сроку' },
	{ code: 'OVERDRAFT', fits: false, reason: 'Только для клиентов банка' },
];
