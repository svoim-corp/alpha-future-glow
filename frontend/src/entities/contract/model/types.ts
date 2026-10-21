export interface ContractTerms {
	productCode: string;
	amount: number;
	term: number;
	rate: number;
	monthlyPayment: number;
	firstPaymentDate: string;
}

export interface ContractDocument {
	id: string;
	title: string;
	url: string;
}

/** GET /applications/{id}/contract */
export interface Contract {
	terms: ContractTerms;
	documents: ContractDocument[];
	signed: boolean;
	disbursed: boolean;
}
