import { planCoversDate } from './period';
import { type BusinessPlan } from './types';

const plan = (patch: Partial<BusinessPlan>): BusinessPlan => ({
	id: 'p-1',
	title: 'План',
	note: null,
	goal: null,
	budgetRub: 100,
	dateMode: 'PERIOD',
	startDate: '2027-06-01',
	endDate: '2027-08-31',
	state: 'ACTIVE',
	notifications: { inApp: true, email: false },
	recommendation: null,
	createdAt: '2026-10-20T09:00:00Z',
	...patch,
});

describe('planCoversDate', () => {
	it('границы периода включены', () => {
		expect(planCoversDate(plan({}), '2027-06-01')).toBe(true);
		expect(planCoversDate(plan({}), '2027-08-31')).toBe(true);
		expect(planCoversDate(plan({}), '2027-09-01')).toBe(false);
		expect(planCoversDate(plan({}), '2027-05-31')).toBe(false);
	});

	it('один день покрывает только себя', () => {
		const day = plan({ dateMode: 'DAY', startDate: '2026-11-05', endDate: '2026-11-05' });

		expect(planCoversDate(day, '2026-11-05')).toBe(true);
		expect(planCoversDate(day, '2026-11-06')).toBe(false);
	});

	it('идея без даты ничего не покрывает', () => {
		expect(planCoversDate(plan({ dateMode: 'UNSCHEDULED', startDate: null, endDate: null }), '2027-06-01')).toBe(false);
	});
});
