import { type BusinessPlan } from './types';

/** Покрывает ли план указанный день (ISO-даты сравниваются как строки) */
export const planCoversDate = (plan: BusinessPlan, date: string): boolean => (
	plan.dateMode !== 'UNSCHEDULED'
	&& plan.startDate !== null
	&& date >= plan.startDate
	&& date <= (plan.endDate ?? plan.startDate)
);
