import { type CalculatePaymentArgs, type PaymentCalculation } from '../api/calculatorApi';

/** Ориентир по округлённым сервером платежам, без повторения формулы кредита на клиенте. */
export const paymentTotals = (params: CalculatePaymentArgs, calculation: PaymentCalculation) => {
	const grace = calculation.gracePayment;
	const months = params.term - (grace?.months ?? 0);
	const min = calculation.payment.min * months + (grace?.min ?? 0) * (grace?.months ?? 0);
	const max = calculation.payment.max * months + (grace?.max ?? 0) * (grace?.months ?? 0);
	return {
		payment: { min, max },
		interest: { min: Math.max(0, min - params.amount), max: Math.max(0, max - params.amount) },
	};
};
