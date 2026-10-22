import { planCoversDate } from './period';
import { type BusinessPlan, type PlanState } from './types';

export const PLAN_STATE_LABELS: Record<PlanState, string> = {
	ACTIVE: 'Активен',
	PAUSED: 'На паузе',
	DONE: 'Завершён',
	ARCHIVED: 'В архиве',
};

/** Сервер Planner использует московский день (фиксированный UTC+3) для демо-компаний. */
export const moscowToday = (now = new Date()): string => (
	new Date(now.getTime() + 3 * 60 * 60 * 1000).toISOString().slice(0, 10)
);

export const plansOnDate = (plans: BusinessPlan[], date: string): BusinessPlan[] => (
	plans.filter((plan) => planCoversDate(plan, date))
);

/** Истечение периода не переводит сохранённое состояние плана в DONE. */
export const planPhase = (plan: BusinessPlan, today: string): string => {
	if (plan.dateMode === 'UNSCHEDULED' || !plan.startDate) {
		return 'Без даты';
	}

	if (plan.startDate > today) {
		return 'Предстоит';
	}

	return (plan.endDate ?? plan.startDate) < today ? 'Период прошёл' : 'Идёт сейчас';
};
