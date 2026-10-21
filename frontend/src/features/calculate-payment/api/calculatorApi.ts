import { type Range } from '@entities/product';
import { baseApi } from '@shared/api';

export interface CalculatePaymentArgs {
	productCode: string;
	amount: number;
	term: number;
	graceMonths?: number | null;
}

export interface GracePayment extends Range {
	months: number;
}

export interface PaymentCalculation {
	rate: Range;
	payment: Range;
	/** Только при отсрочке: проценты без основного долга */
	gracePayment: GracePayment | null;
	withinProductLimits: boolean;
}

export const calculatorApi = baseApi.injectEndpoints({
	endpoints: (build) => ({
		/** Расчёт платежа — POST без побочных эффектов, поэтому query: результат кэшируется по параметрам */
		calculatePayment: build.query<PaymentCalculation, CalculatePaymentArgs>({
			query: (body) => ({ url: '/calculator/payment', method: 'POST', body }),
		}),
	}),
});

export const { useCalculatePaymentQuery } = calculatorApi;
