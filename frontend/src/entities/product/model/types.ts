export interface Range {
	min: number;
	max: number;
}

/** Карточка каталога: GET /products */
export interface ProductSummary {
	code: string;
	name: string;
	shortText: string;
	goals: string[];
	amount: Range;
	/** У овердрафта срока кредита нет */
	term?: Range | null;
	rate: Range;
	noStatementThreshold: number | null;
	bankClientsOnly: boolean;
}

export interface ProductRequirements {
	minBusinessAgeMonths?: number | null;
	bankClientsOnly: boolean;
	guarantorForLlc: boolean;
}

/** Условия продукта: GET /products/{code} */
export interface ProductDetails {
	code: string;
	name: string;
	description: string;
	goals: string[];
	amount: Range;
	term?: Range | null;
	rate: Range;
	graceMaxMonths?: number | null;
	requirements: ProductRequirements;
	/** null — выписка нужна всегда */
	noStatementThreshold: number | null;
	limitRule?: string | null;
	whatYouNeed: string[];
}

export interface Goal {
	code: string;
	name: string;
}

export interface MatchRequest {
	goal: string;
	amount: number;
	term: number;
}

export interface ProductMatch {
	code: string;
	fits: boolean;
	reason: string;
}
