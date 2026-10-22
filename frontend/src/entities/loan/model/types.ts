export interface PaymentRef {
	date: string;
	amount: number;
}

/** GET /loans/summary: есть ли выданные кредиты и ближайший платёж */
export interface LoansSummary {
	activeLoans: number;
	totalPrincipalBalance?: number;
	closedLoans?: number;
	totalLoans?: number;
	overdue?: { loans: number; payments: number; amount: number };
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
	id?: string;
	date: string;
	principal: number;
	interest: number;
	total: number;
	/** Остаток к оплате с учётом частичных платежей; новый контракт Servicing. */
	remainingTotal?: number;
	paidPrincipal?: number;
	paidInterest?: number;
	remainingPrincipal?: number;
	remainingInterest?: number;
	early?: boolean;
	status: PaymentStatus;
}

export interface LoanTerms {
	amount: number;
	term: number;
	rate: number;
	monthlyPayment: number;
	graceMonths?: number;
}

/** GET /loans/{id} */
export interface LoanDetails {
	id: string;
	applicationId: string;
	contractNumber: string;
	product?: LoanProductRef;
	balance?: number;
	issuedAt?: string;
	status?: LoanStatus;
	revolving?: boolean;
	closedAt?: string | null;
	closeReason?: 'SCHEDULE_COMPLETED' | 'EARLY_REPAYMENT' | 'LIMIT_CLOSED' | null;
	version?: number;
	scheduleVersion?: number;
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

export interface LoanDebt {
	loanId: string;
	loanVersion: number;
	scheduleVersion: number;
	asOf: string;
	businessDate: string;
	timeZone: string;
	principalOutstanding: number;
	interestOutstanding: number;
	penaltiesOutstanding: number;
	feesOutstanding: number;
	totalOutstanding: number;
	prepaidInterest: number;
	unappliedFunds: number;
	calculationPolicyVersion: string;
	overdue: {
		since: string | null;
		days: number;
		principal: number;
		interest: number;
		penalties: number;
		fees: number;
		total: number;
	};
}

export type RepaymentStatus = 'PENDING' | 'PROCESSING' | 'SUCCEEDED' | 'FAILED';

export interface LoanRepayment {
	id: string;
	loanId: string;
	type: 'REGULAR' | 'EARLY';
	status: RepaymentStatus;
	account: { id: string; maskedNumber: string };
	requestedAmount: number;
	postedAmount: number | null;
	allocation: { principal: number; interest: number; penalties: number; fees: number } | null;
	unappliedAmount: number | null;
	externalTransactionId: string | null;
	failureReason: string | null;
	quoteId: string | null;
	requestedAt: string;
	postedAt: string | null;
	loanVersion: number;
	scheduleVersion: number;
	loanStatus: LoanStatus;
}
