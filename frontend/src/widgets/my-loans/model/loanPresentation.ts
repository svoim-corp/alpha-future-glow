import { isExpiredOffer, type PendingOffer } from '@entities/application';
import { type Loan, type LoanPayment, paymentAmounts } from '@entities/loan';
import { formatDate } from '@shared/lib/format';

/** Сколько ближайших платежей показываем на графике и в таблице (макет loans.html) */
export const UPCOMING_COUNT = 6;

export const pendingOfferTotal = (offers: PendingOffer[], today: string) =>
	offers.reduce((total, entry) => total + (isExpiredOffer(entry, today) ? 0 : entry.offer?.offered.amount ?? 0), 0);

/** Доля погашенного основного долга, % */
export const repaidPercent = (loan: Loan) =>
	loan.amount > 0 ? Math.min(100, Math.max(0, Math.round(((loan.amount - loan.balance) / loan.amount) * 100))) : 0;

export const upcomingPayments = (payments: LoanPayment[]) =>
	payments.filter((payment) => payment.status === 'PLANNED').slice(0, UPCOMING_COUNT);

export const remainingCount = (payments: LoanPayment[]) => payments.filter((payment) => payment.status !== 'PAID').length;

/** «ноя» — подпись столбца графика */
export const monthLabel = (iso: string) => {
	const label = new Intl.DateTimeFormat('ru-RU', { month: 'short' }).format(new Date(`${iso}T00:00:00`)).slice(0, 3);

	return label.charAt(0).toUpperCase() + label.slice(1);
};

/** «15 ноября» и «2026» для блока ближайшего платежа */
export const paymentDay = (iso: string) => {
	const date = new Date(`${iso}T00:00:00`);

	return {
		day: new Intl.DateTimeFormat('ru-RU', { day: 'numeric', month: 'long' }).format(date),
		year: String(date.getFullYear()),
	};
};

/** Верх шкалы графика: ближайшая «круглая» отметка выше максимального платежа (кратно 20 тыс.) */
export const chartCeiling = (payments: LoanPayment[]) =>
	Math.max(20000, Math.ceil(Math.max(0, ...payments.map((payment) => payment.total)) / 20000) * 20000);

export const thousands = (value: number) => `${new Intl.NumberFormat('ru-RU', { maximumFractionDigits: 1 }).format(value / 1000)} тыс.`;

export const scheduleCsv = (payments: LoanPayment[]) => {
	const labels = { PLANNED: 'Запланирован', PAID: 'Оплачен', OVERDUE: 'Просрочен' };
	return '\uFEFF' + [
		'Дата;Основной долг по графику, ₽;Проценты по графику, ₽;Всего по графику, ₽;Оплачено, ₽;Осталось, ₽;Статус;Досрочное погашение',
		...payments.map((payment) => [formatDate(payment.date), payment.principal, payment.interest, payment.total, paymentAmounts(payment).paid, paymentAmounts(payment).remaining, labels[payment.status], payment.early ? 'Да' : 'Нет'].join(';')),
	].join('\r\n');
};
