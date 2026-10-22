import { intersects } from '@shared/lib/grid';
import { z } from 'zod';

import { DASHBOARD_DEFINITIONS } from './definitions';

const reference = z.string().trim().min(1).max(128).regex(/^[A-Za-z0-9_-]+$/u).nullable();
const base = z.object({
	id: z.string().min(1).max(64),
	x: z.number().int().min(0).max(11), y: z.number().int().min(0).max(200),
	w: z.number().int().min(1).max(12), h: z.number().int().min(2).max(5),
}).strict();
export const quickActionSchema = z.enum(['apply', 'plan', 'credits', 'calendar', 'support']);
export const dashboardWidgetSchema = z.discriminatedUnion('type', [
	base.extend({ type: z.literal('applications'), settings: z.object({ filter: z.enum(['all', 'attention']) }).strict() }),
	base.extend({ type: z.literal('loan'), settings: z.object({ loanId: reference }).strict() }),
	base.extend({ type: z.literal('payment'), settings: z.object({ loanId: reference }).strict() }),
	base.extend({ type: z.literal('debt'), settings: z.object({}).strict() }),
	base.extend({ type: z.literal('quick'), settings: z.object({ actions: z.array(quickActionSchema).min(1).max(5).refine((actions) => new Set(actions).size === actions.length) }).strict() }),
	base.extend({ type: z.literal('calendar'), settings: z.object({ source: z.enum(['all', 'payments', 'plans']) }).strict() }),
	base.extend({ type: z.literal('plans'), settings: z.object({ filter: z.enum(['upcoming', 'ideas']) }).strict() }),
	base.extend({ type: z.literal('application'), settings: z.object({ applicationId: reference }).strict() }),
	base.extend({ type: z.literal('loans'), settings: z.object({ status: z.enum(['ACTIVE', 'CLOSED']) }).strict() }),
	base.extend({ type: z.literal('payments'), settings: z.object({ days: z.union([z.literal(7), z.literal(30)]), loanId: reference }).strict() }),
	base.extend({ type: z.literal('offers'), settings: z.object({ filter: z.enum(['current', 'all']) }).strict() }),
	base.extend({ type: z.literal('drafts'), settings: z.object({}).strict() }),
	base.extend({ type: z.literal('planFunding'), settings: z.object({ planId: reference }).strict() }),
	base.extend({ type: z.literal('notifications'), settings: z.object({ filter: z.enum(['all', 'unread']) }).strict() }),
	base.extend({ type: z.literal('tickets'), settings: z.object({ filter: z.enum(['active', 'answered', 'all']) }).strict() }),
	base.extend({ type: z.literal('funding'), settings: z.object({}).strict() }),
]);
export const dashboardConfigSchema = z.object({
	schemaVersion: z.literal(1),
	widgets: z.array(dashboardWidgetSchema).max(30),
}).strict().superRefine(({ widgets }, ctx) => {
	const ids = new Set<string>();
	widgets.forEach((item, index) => {
		if (ids.has(item.id)) ctx.addIssue({ code: 'custom', path: ['widgets', index, 'id'], message: 'Повторяющийся идентификатор виджета' });
		ids.add(item.id);
		if (item.x + item.w > 12 || !DASHBOARD_DEFINITIONS[item.type].sizes.some(([w, h]) => w === item.w && h === item.h)) {
			ctx.addIssue({ code: 'custom', path: ['widgets', index], message: 'Недопустимый размер или положение виджета' });
		}
		if (widgets.slice(0, index).some((other) => intersects(item, other))) {
			ctx.addIssue({ code: 'custom', path: ['widgets', index], message: 'Виджеты не должны пересекаться' });
		}
	});
});
export type DashboardConfig = z.infer<typeof dashboardConfigSchema>;
export type DashboardWidget = z.infer<typeof dashboardWidgetSchema>;
export type DashboardWidgetKind = DashboardWidget['type'];
export type QuickAction = z.infer<typeof quickActionSchema>;
