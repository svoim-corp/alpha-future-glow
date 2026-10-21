import { makePlan } from '@test/fixtures/plans';
import { apiUrl, server } from '@test/msw/server';

import { db, session } from './db';
import { plannerHandlers } from './planner';

const originalPlans = structuredClone(db.plans);
beforeEach(() => {
	session.set('CLIENT');
	db.plans = [makePlan()];
	server.use(...plannerHandlers);
});
afterEach(() => {
	db.plans = structuredClone(originalPlans);
	session.set(null);
});
const post = (body: unknown) => fetch(apiUrl('/plans'), { method: 'POST', body: JSON.stringify(body), headers: { 'Content-Type': 'application/json' } });
const patch = (body: unknown) => fetch(apiUrl('/plans/p-today'), { method: 'PATCH', body: JSON.stringify(body), headers: { 'Content-Type': 'application/json' } });
const input = { title: 'Печь', note: null, goal: 'EQUIPMENT', budgetRub: 350000,
	dateMode: 'PERIOD', startDate: '2026-10-20', endDate: '2026-10-23', notifications: { inApp: true, email: false } };

describe('демо API планера', () => {
	it('сохраняет период и возвращает полный план, а не заглушку', async () => {
		const response = await post(input);
		expect(await response.clone().text()).not.toContain('TypeError');
		expect(response.status).toBe(201);
		expect(await response.json()).toMatchObject({ ...input, state: 'ACTIVE', recommendation: null });
	});

	it('проверяет цель и даты вместо принятия произвольного запроса', async () => {
		expect((await post({ ...input, goal: null })).status).toBe(400);
		expect((await post({ ...input, endDate: '2026-10-19' })).status).toBe(400);
		expect((await post({ ...input, startDate: '2026-02-30' })).status).toBe(400);
	});

	it('PATCH нормализует даты и очищает заметку по серверным правилам', async () => {
		const day = await patch({ dateMode: 'DAY', startDate: '2026-10-21' });
		expect(await day.json()).toMatchObject({ startDate: '2026-10-21', endDate: '2026-10-21' });
		const idea = await patch({ dateMode: 'UNSCHEDULED', note: '' });
		expect(await idea.json()).toMatchObject({ startDate: null, endDate: null, note: null });
	});

	it('архив исключается из общего списка и возвращается по фильтру; сводка совпадает с набором', async () => {
		await patch({ state: 'ARCHIVED' });
		const regular = await fetch(apiUrl('/plans'));
		expect(await regular.json()).toMatchObject({ items: [], summary: { total: 0, budgetRub: 0 } });
		const archived = await fetch(apiUrl('/plans?state=ARCHIVED'));
		expect(await archived.json()).toMatchObject({ items: [{ state: 'ARCHIVED' }], summary: { total: 1 } });
	});

	it('не выдумывает подбор для новой идеи без даты', async () => {
		await post({ ...input, dateMode: 'UNSCHEDULED', startDate: null, endDate: null });
		const response = await fetch(apiUrl('/plans/recommendations'));
		expect(await response.json()).toEqual([]);
	});

	it('удаляет существующий план, а отсутствующий возвращает 404', async () => {
		expect((await fetch(apiUrl('/plans/p-today'), { method: 'DELETE' })).status).toBe(204);
		expect((await patch({ state: 'DONE' })).status).toBe(404);
		expect((await fetch(apiUrl('/plans/p-today'), { method: 'DELETE' })).status).toBe(404);
	});

	it('отклоняет запрос без сессии и без компании', async () => {
		session.set(null);
		expect((await fetch(apiUrl('/plans'))).status).toBe(401);
		session.set('BANK_OFFICER');
		expect((await fetch(apiUrl('/plans'))).status).toBe(403);
	});
});
