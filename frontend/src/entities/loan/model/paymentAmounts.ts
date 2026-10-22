import { type LoanPayment } from './types';

/** Новые поля сервера приоритетны; для прежнего контракта остаётся смысл статуса PAID. */
export const paymentAmounts = (payment: LoanPayment) => {
	const paidPrincipal = payment.paidPrincipal ?? (payment.status === 'PAID' ? payment.principal : 0);
	const paidInterest = payment.paidInterest ?? (payment.status === 'PAID' ? payment.interest : 0);
	return {
		paid: paidPrincipal + paidInterest,
		principal: payment.remainingPrincipal ?? Math.max(0, payment.principal - paidPrincipal),
		interest: payment.remainingInterest ?? Math.max(0, payment.interest - paidInterest),
		remaining: payment.remainingTotal ?? Math.max(0, payment.total - paidPrincipal - paidInterest),
	};
};
