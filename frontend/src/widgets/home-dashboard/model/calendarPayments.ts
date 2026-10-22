import { type ScheduleItem } from '@entities/loan';

/** API графика не отдаёт ID строки. Одинаковые события объединяем без потери суммы и количества. */
export const calendarPayments = (items: readonly ScheduleItem[]) => {
	const groups = new Map<string, ScheduleItem & { key: string; count: number }>();
	for (const item of items) {
		const key = JSON.stringify([item.loanId, item.date, item.status]);
		const previous = groups.get(key);
		groups.set(key, {
			...item, key,
			amount: (previous?.amount ?? 0) + item.amount,
			count: (previous?.count ?? 0) + 1,
		});
	}
	return [...groups.values()];
};
