import { type Contract } from '@entities/contract';

export const contractFixture: Contract = {
	terms: {
		productCode: 'WORKING',
		amount: 1200000,
		term: 24,
		rate: 21.5,
		monthlyPayment: 62200,
		firstPaymentDate: '2026-11-20',
	},
	documents: [
		{ id: 'd-1', title: 'Кредитный договор', url: '/api/v1/documents/d-1' },
		{ id: 'd-2', title: 'Договор поручительства', url: '/api/v1/documents/d-2' },
	],
	signed: false,
	disbursed: false,
};
