import { type LoanDetails, type ScheduleItem } from './types';

export const loanOverview = (loan: LoanDetails) => ({
	closed: loan.status === 'CLOSED',
	percent: loan.balance !== undefined && loan.terms.amount > 0
		? Math.min(100, Math.max(0, Math.round((1 - loan.balance / loan.terms.amount) * 100))) : null,
	next: loan.status === 'CLOSED' ? undefined : [...loan.payments].filter((payment) => payment.status === 'PLANNED').sort((a, b) => a.date.localeCompare(b.date))[0],
	overdue: loan.payments.filter((payment) => payment.status === 'OVERDUE').length,
});

/** Скользящий период включает сегодня и ещё days - 1 календарных дней по UTC. */
export const upcomingPaymentRange = (days: 7 | 30, now = new Date()) => {
	const from = now.toISOString().slice(0, 10);
	const end = new Date(`${from}T00:00:00Z`);
	end.setUTCDate(end.getUTCDate() + days - 1);
	return { from, to: end.toISOString().slice(0, 10) };
};

/** Только будущие запланированные строки, сгруппированные по дате, без смешения с оплатами. */
export const paymentGroups = (items: readonly ScheduleItem[], loanId: string | null) => {
	const groups = new Map<string, { date: string; amount: number; count: number }>();
	for (const item of items) {
		if (item.status !== 'PLANNED' || (loanId !== null && item.loanId !== loanId)) continue;
		const current = groups.get(item.date) ?? { date: item.date, amount: 0, count: 0 };
		groups.set(item.date, { date: item.date, amount: current.amount + item.amount, count: current.count + 1 });
	}
	return [...groups.values()].sort((a, b) => a.date.localeCompare(b.date));
};
