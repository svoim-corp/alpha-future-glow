import { type BusinessPlan } from './types';

/** Покрывает ли план указанный день (ISO-даты сравниваются как строки) */
export const planCoversDate = (plan: BusinessPlan, date: string): boolean => (
	plan.dateMode !== 'UNSCHEDULED'
	&& plan.startDate !== null
	&& date >= plan.startDate
	&& date <= (plan.endDate ?? plan.startDate)
);

const dayMonth = new Intl.DateTimeFormat('ru-RU', { day: 'numeric', month: 'long' });
const dayMonthYear = new Intl.DateTimeFormat('ru-RU', { day: 'numeric', month: 'long', year: 'numeric' });
const clean = (text: string) => text.replace(/\s*г\.$/u, '');
const parse = (iso: string) => new Date(`${iso}T00:00:00`);

/** «1 июня — 31 августа 2027»; день — «5 июня 2027»; идея без даты — «Без даты» */
export const formatPeriodLong = (plan: BusinessPlan): string => {
	if (plan.dateMode === 'UNSCHEDULED' || !plan.startDate) {
		return 'Без даты';
	}

	const start = parse(plan.startDate);

	if (!plan.endDate || plan.endDate === plan.startDate) {
		return clean(dayMonthYear.format(start));
	}

	const end = parse(plan.endDate);
	const from = start.getFullYear() === end.getFullYear() ? dayMonth.format(start) : clean(dayMonthYear.format(start));

	return `${from} — ${clean(dayMonthYear.format(end))}`;
};

/** «1 июня 2027» — дата начала плана */
export const formatPlanStart = (plan: BusinessPlan): string => (
	plan.startDate ? clean(dayMonthYear.format(parse(plan.startDate))) : 'сразу'
);
