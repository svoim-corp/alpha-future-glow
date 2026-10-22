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
	/** Сервер возвращает диапазон; у овердрафта min = max = 12. */
	term?: Range | null;
	rate: Range;
}

export interface ProductRequirements {
	minBusinessAgeMonths?: number | null;
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
