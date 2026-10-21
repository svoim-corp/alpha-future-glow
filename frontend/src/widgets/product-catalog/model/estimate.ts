import { type CalculatePaymentArgs } from '@features/calculate-payment';

/** Параметры расчёта платежа; пока сумма или срок не введены корректно — null, запрос не уходит */
export const estimateParams = (
	productCode: string,
	amount: number | null,
	term: number | null,
): CalculatePaymentArgs | null => (
	amount !== null && term !== null ? { productCode, amount, term } : null
);
