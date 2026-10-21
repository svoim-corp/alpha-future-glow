import { type PlansResponse } from '@entities/plan';

export const plansFixture: PlansResponse = {
	items: [
		{
			id: 'p-1',
			title: 'Ремонт помещения',
			note: 'Обновить торговый зал',
			goal: 'REAL_ESTATE',
			budgetRub: 2400000,
			dateMode: 'PERIOD',
			startDate: '2027-06-01',
			endDate: '2027-08-31',
			state: 'ACTIVE',
			notifications: { inApp: true, email: true },
			recommendation: null,
			createdAt: '2026-10-20T09:00:00Z',
		},
		{
			id: 'p-2',
			title: 'Запасы к зимнему сезону',
			note: null,
			goal: 'WORKING_CAPITAL',
			budgetRub: 650000,
			dateMode: 'PERIOD',
			startDate: '2026-11-01',
			endDate: '2026-11-30',
			state: 'ACTIVE',
			notifications: { inApp: true, email: false },
			recommendation: null,
			createdAt: '2026-10-20T09:00:00Z',
		},
		{
			id: 'p-3',
			title: 'Новая печь для пекарни',
			note: 'Увеличить выпуск хлеба',
			goal: 'EQUIPMENT',
			budgetRub: 1200000,
			dateMode: 'UNSCHEDULED',
			startDate: null,
			endDate: null,
			state: 'ACTIVE',
			notifications: { inApp: true, email: false },
			recommendation: {
				productCode: 'INVEST',
				productName: 'Инвестиционный кредит',
				reason: 'Цель и сумма укладываются в рамки продукта',
				availableFrom: '2026-10-20',
			},
			createdAt: '2026-10-20T09:00:00Z',
		},
	],
	summary: { total: 3, withPeriod: 2, unscheduled: 1, budgetRub: 4250000 },
};

export const makePlan = (patch: Partial<PlansResponse['items'][number]> = {}): PlansResponse['items'][number] => ({
	id: 'p-today', title: 'Закупка оборудования', note: null, goal: 'EQUIPMENT', budgetRub: 1200000,
	dateMode: 'PERIOD', startDate: '2026-10-20', endDate: '2026-10-22', state: 'ACTIVE',
	notifications: { inApp: true, email: false }, recommendation: null, createdAt: '2026-10-20T09:00:00Z',
	...patch,
});
