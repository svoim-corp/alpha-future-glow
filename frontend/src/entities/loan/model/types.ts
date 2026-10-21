export interface PaymentRef {
	date: string;
	amount: number;
}

/** GET /loans/summary: есть ли выданные кредиты и ближайший платёж */
export interface LoansSummary {
	activeLoans: number;
	nextPayment: PaymentRef | null;
}

export type LoanStatus = 'ACTIVE' | 'CLOSED';

export interface LoanProductRef {
	code: string;
	name: string;
}

/** Строка списка GET /loans */
export interface Loan {
	id: string;
	product: LoanProductRef;
	amount: number;
	balance: number;
	rate: number;
	term: number;
	issuedAt: string;
	nextPayment: PaymentRef | null;
	status: LoanStatus;
}

export type PaymentStatus = 'PLANNED' | 'PAID' | 'OVERDUE';

export interface LoanPayment {
	date: string;
	principal: number;
	interest: number;
	total: number;
	status: PaymentStatus;
}

export interface LoanTerms {
	amount: number;
	term: number;
	rate: number;
	monthlyPayment: number;
}

/** GET /loans/{id} */
export interface LoanDetails {
	id: string;
	applicationId: string;
	contractNumber: string;
	terms: LoanTerms;
	payments: LoanPayment[];
}

/** Элемент GET /loans/schedule и /loans/schedule/upcoming */
export interface ScheduleItem {
	date: string;
	loanId: string;
	productName: string;
	amount: number;
	status: PaymentStatus;
}
