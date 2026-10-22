import { plansFixture } from '@test/fixtures/plans';

import { moscowToday, planPhase, plansOnDate } from './calendar';
import { type BusinessPlan } from './types';

const dated: BusinessPlan = {
	id: 'day', title: 'Закупка', note: null, goal: 'EQUIPMENT', budgetRub: 100,
	dateMode: 'PERIOD', startDate: '2026-10-20', endDate: '2026-10-22', state: 'ACTIVE',
	notifications: { inApp: true, email: false }, recommendation: null, createdAt: '2026-10-20T09:00:00Z',
};

describe('календарные правила Planner', () => {
	it('сегодня определяется по московскому дню, включая переход года', () => {
		expect(moscowToday(new Date('2026-10-20T20:59:59Z'))).toBe('2026-10-20');
		expect(moscowToday(new Date('2026-10-20T21:00:00Z'))).toBe('2026-10-21');
		expect(moscowToday(new Date('2026-12-31T22:00:00Z'))).toBe('2027-01-01');
	});

	it('возвращает все пересекающиеся планы, включая границы периода и неактивные состояния', () => {
		const another = { ...dated, id: 'another', state: 'PAUSED' as const };
		expect(plansOnDate([dated, another, ...plansFixture.items], '2026-10-20')).toEqual([dated, another]);
		expect(plansOnDate([dated], '2026-10-22')).toEqual([dated]);
		expect(plansOnDate([dated], '2026-10-23')).toEqual([]);
	});

	it('не помещает идею без даты в календарь', () => {
		expect(plansOnDate([{ ...dated, dateMode: 'UNSCHEDULED', startDate: null, endDate: null }], '2026-10-21')).toEqual([]);
	});

	it('различает будущий, текущий, прошедший период и идею, не меняя состояние', () => {
		expect(planPhase(dated, '2026-10-19')).toBe('Предстоит');
		expect(planPhase(dated, '2026-10-20')).toBe('Идёт сейчас');
		expect(planPhase(dated, '2026-10-22')).toBe('Идёт сейчас');
		expect(planPhase(dated, '2026-10-23')).toBe('Период прошёл');
		expect(dated.state).toBe('ACTIVE');
		expect(planPhase({ ...dated, dateMode: 'UNSCHEDULED', startDate: null }, '2026-10-21')).toBe('Без даты');
	});
});
