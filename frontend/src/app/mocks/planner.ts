import { type BusinessPlan, moscowToday, type PlansResponse } from '@entities/plan';
import { env } from '@shared/config';
import { http, HttpResponse } from 'msw';
import { z } from 'zod';

import { db, nextId, session } from './db';

const url = (path: string) => `${env.apiBaseUrl}${path}`;
const states = ['ACTIVE', 'PAUSED', 'DONE', 'ARCHIVED'] as const;
const date = z.iso.date({ error: 'Укажите корректную дату' }).nullable().optional();
const notifications = z.object({ inApp: z.boolean().optional(), email: z.boolean().optional() }).nullable().optional();
const fields = {
	title: z.string({ error: 'Укажите название' }).min(1, 'Назовите план').max(200, 'Не больше 200 символов'), note: z.string({ error: 'Проверьте заметку' }).max(2000, 'Не больше 2000 символов').nullable().optional(),
	goal: z.string({ error: 'Выберите цель' }).min(1, 'Выберите цель').max(64, 'Не больше 64 символов'), budgetRub: z.number({ error: 'Укажите бюджет' }).int('Бюджет — целое число рублей')
		.positive('Бюджет должен быть больше нуля').max(Number.MAX_SAFE_INTEGER, 'Сумма слишком велика'),
	dateMode: z.enum(['DAY', 'PERIOD', 'UNSCHEDULED'], { error: 'Выберите режим даты' }), startDate: date, endDate: date, notifications,
};
const createSchema = z.object(fields);
const patchSchema = z.object({
	...createSchema.partial().shape, state: z.enum(states, { error: 'Недопустимое состояние' }).optional(),
});
const failure = (status: number, code: string, message: string) => HttpResponse.json({ code, message, traceId: 'planner-demo' }, { status });
const access = () => {
	const user = session.get();
	if (!user) return failure(401, 'UNAUTHORIZED', 'Требуется вход в систему');
	if (!user.company) return failure(403, 'FORBIDDEN', 'Планер доступен только клиенту');

	return null;
};
const validation = (issues: { path: PropertyKey[]; message: string }[]) => HttpResponse.json({
	code: 'VALIDATION_ERROR', message: 'Проверьте поля плана', traceId: 'planner-demo',
	fields: issues.map((issue) => ({ field: issue.path.join('.'), code: 'INVALID', message: issue.message })),
}, { status: 400 });

const normalise = (plan: BusinessPlan): BusinessPlan => {
	if (plan.dateMode === 'UNSCHEDULED') return { ...plan, startDate: null, endDate: null };
	if (plan.dateMode === 'DAY') return { ...plan, endDate: plan.startDate };

	return plan;
};
const dateError = (plan: BusinessPlan) => {
	if (plan.dateMode !== 'UNSCHEDULED' && !plan.startDate) return validation([{ path: ['startDate'], message: 'Укажите дату начала' }]);
	if (plan.dateMode === 'PERIOD' && !plan.endDate) return validation([{ path: ['endDate'], message: 'Укажите дату окончания' }]);
	if (plan.startDate && plan.endDate && plan.endDate < plan.startDate) return validation([{ path: ['endDate'], message: 'Дата окончания не может быть раньше даты начала' }]);

	return null;
};
const view = (plan: BusinessPlan): BusinessPlan => ({
	...plan,
	recommendation: plan.state !== 'ACTIVE' || (plan.endDate && plan.endDate < moscowToday()) ? null : plan.recommendation,
});
const response = (items: BusinessPlan[]): PlansResponse => ({
	items: items.map(view), summary: {
		total: items.length, withPeriod: items.filter((p) => p.dateMode !== 'UNSCHEDULED').length,
		unscheduled: items.filter((p) => p.dateMode === 'UNSCHEDULED').length,
		budgetRub: items.reduce((sum, p) => sum + p.budgetRub, 0),
	},
});

/** Демо повторяет операции Planner. Новый подбор без ответа Catalog не выдумывается. */
export const plannerHandlers = [
	http.get(url('/plans/recommendations'), () => {
		const denied = access();
		if (denied) return denied;

		return HttpResponse.json(db.plans.map(view).flatMap((plan) => {
			const rec = plan.recommendation;
			if (!rec || rec.availableFrom > moscowToday()) return [];

			return [{
				planId: plan.id, planTitle: plan.title, budgetRub: plan.budgetRub,
				productCode: rec.productCode, productName: rec.productName, reason: rec.reason,
			}];
		}));
	}),
	http.get(url('/plans'), ({ request }) => {
		const denied = access();
		if (denied) return denied;
		const state = new URL(request.url).searchParams.get('state');
		if (state && !states.some((s) => s === state)) return validation([{ path: ['state'], message: 'Недопустимое состояние' }]);
		const items = db.plans.filter((p) => state ? p.state === state : p.state !== 'ARCHIVED')
			.toSorted((a, b) => b.createdAt.localeCompare(a.createdAt) || a.id.localeCompare(b.id)).slice(0, 200);

		return HttpResponse.json(response(items));
	}),
	http.post(url('/plans'), async ({ request }) => {
		const denied = access();
		if (denied) return denied;
		const result = createSchema.safeParse(await request.json().catch(() => null));
		if (!result.success) return validation(result.error.issues);
		const data = result.data;
		const plan = normalise({
			...data, id: nextId('p'), note: data.note || null, startDate: data.startDate ?? null, endDate: data.endDate ?? null,
			state: 'ACTIVE', notifications: { inApp: data.notifications?.inApp ?? true, email: data.notifications?.email ?? false },
			recommendation: null, createdAt: new Date().toISOString(),
		});
		const error = dateError(plan);
		if (error) return error;
		db.plans.unshift(plan);

		return HttpResponse.json(plan, { status: 201 });
	}),
	http.patch(url('/plans/:id'), async ({ params, request }) => {
		const denied = access();
		if (denied) return denied;
		const plan = db.plans.find((p) => p.id === params.id);
		if (!plan) return failure(404, 'NOT_FOUND', 'План не найден');
		const result = patchSchema.safeParse(await request.json().catch(() => null));
		if (!result.success) return validation(result.error.issues);
		const data = result.data;
		const updated = normalise({
			...plan, title: data.title ?? plan.title, goal: data.goal ?? plan.goal,
			budgetRub: data.budgetRub ?? plan.budgetRub,
			note: data.note === '' ? null : data.note ?? plan.note,
			dateMode: data.dateMode ?? plan.dateMode,
			state: data.state ?? plan.state,
			startDate: data.startDate ?? plan.startDate, endDate: data.endDate ?? plan.endDate,
			notifications: { ...plan.notifications, ...data.notifications },
			recommendation: [data.goal, data.budgetRub, data.dateMode, data.startDate, data.endDate]
				.some((value) => value !== undefined) ? null : plan.recommendation,
		});
		const error = dateError(updated);
		if (error) return error;
		db.plans = db.plans.map((p) => p.id === updated.id ? updated : p);

		return HttpResponse.json(view(updated));
	}),
	http.delete(url('/plans/:id'), ({ params }) => {
		const denied = access();
		if (denied) return denied;
		if (!db.plans.some((p) => p.id === params.id)) return failure(404, 'NOT_FOUND', 'План не найден');
		db.plans = db.plans.filter((p) => p.id !== params.id);

		return new HttpResponse(null, { status: 204 });
	}),
];
