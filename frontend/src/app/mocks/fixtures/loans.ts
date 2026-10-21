import { type Loan, type LoanDetails, type ScheduleItem } from '@entities/loan';

export const loanFixture: Loan = {
	id: 'l-55',
	product: { code: 'WORKING', name: 'Оборотный кредит' },
	amount: 1200000,
	balance: 1200000,
	rate: 21.5,
	term: 24,
	issuedAt: '2026-10-20',
	nextPayment: { date: '2026-11-20', amount: 62200 },
	status: 'ACTIVE',
};

export const loanDetailsFixture: LoanDetails = {
	id: 'l-55',
	applicationId: 'a-1031',
	contractNumber: 'КД-2026-0055',
	terms: { amount: 1200000, term: 24, rate: 21.5, monthlyPayment: 62200 },
	payments: [
		{ date: '2026-11-20', principal: 40700, interest: 21500, total: 62200, status: 'PLANNED' },
		{ date: '2026-12-20', principal: 41400, interest: 20800, total: 62200, status: 'PLANNED' },
	],
};

export const scheduleFixture: ScheduleItem[] = [
	{ date: '2026-11-20', loanId: 'l-55', productName: 'Оборотный кредит', amount: 62200, status: 'PLANNED' },
	{ date: '2026-12-20', loanId: 'l-55', productName: 'Оборотный кредит', amount: 62200, status: 'PLANNED' },
];
