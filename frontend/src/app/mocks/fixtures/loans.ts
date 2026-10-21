import { type Loan, type LoanDetails, type LoanPayment, type ScheduleItem } from '@entities/loan';

const PAYMENT_COUNT = 18;
const MONTHLY = 77800;

/** Аннуитетный график, как в макете loans.html: платёж 77 800 ₽, доля долга растёт на 1 000 ₽ в месяц */
const payments: LoanPayment[] = Array.from({ length: PAYMENT_COUNT }, (_, index) => {
	const principal = 55000 + index * 1000;
	const date = new Date(Date.UTC(2026, 10 + index, 15));

	return {
		date: date.toISOString().slice(0, 10),
		principal,
		interest: MONTHLY - principal,
		total: MONTHLY,
		status: 'PLANNED',
	};
});

export const loanFixture: Loan = {
	id: 'l-55',
	product: { code: 'WORKING', name: 'Оборотный кредит' },
	amount: 1500000,
	balance: 1160000,
	rate: 22.5,
	term: 24,
	issuedAt: '2026-04-15',
	nextPayment: { date: '2026-11-15', amount: MONTHLY },
	status: 'ACTIVE',
};

export const loanDetailsFixture: LoanDetails = {
	id: 'l-55',
	applicationId: 'a-1031',
	contractNumber: 'КР-2026-000618',
	terms: { amount: 1500000, term: 24, rate: 22.5, monthlyPayment: MONTHLY },
	payments,
};

export const scheduleFixture: ScheduleItem[] = payments.map((payment) => ({
	date: payment.date,
	loanId: 'l-55',
	productName: 'Оборотный кредит',
	amount: payment.total,
	status: payment.status,
}));
