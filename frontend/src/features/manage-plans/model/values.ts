import { type BusinessPlan, type PlanPatch } from '@entities/plan';

import { type PlanValues } from './schema';

export const initialValues = (plan?: BusinessPlan, date?: string): PlanValues => ({
	title: plan?.title ?? '', budget: plan ? String(plan.budgetRub) : '', goal: plan?.goal ?? '',
	dateMode: plan?.dateMode ?? (date ? 'DAY' : 'UNSCHEDULED'),
	startDate: plan?.startDate ?? date ?? '', endDate: plan?.endDate ?? date ?? '',
	note: plan?.note ?? '', inApp: plan?.notifications.inApp ?? true,
});

/** Пустые даты не отправляем в PATCH: сервер очищает их при переходе на UNSCHEDULED. */
export const changedFields = (values: PlanValues, plan: BusinessPlan): PlanPatch => {
	const body: PlanPatch = {};
	if (values.title !== plan.title) body.title = values.title;
	if (values.note !== (plan.note ?? '')) body.note = values.note;
	if (values.goal !== plan.goal) body.goal = values.goal;
	if (Number(values.budget) !== plan.budgetRub) body.budgetRub = Number(values.budget);
	if (values.dateMode !== plan.dateMode) body.dateMode = values.dateMode;
	if (values.dateMode !== 'UNSCHEDULED') {
		if (values.startDate !== plan.startDate) body.startDate = values.startDate;
		if (values.dateMode === 'PERIOD' && values.endDate !== plan.endDate) body.endDate = values.endDate;
	}
	if (values.inApp !== plan.notifications.inApp) body.notifications = { inApp: values.inApp };

	return body;
};
