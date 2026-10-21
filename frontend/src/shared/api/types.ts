/** Ответ со списком и пагинацией (docs/API: «Соглашения API») */
export interface Paginated<T> {
	items: T[];
	page: number;
	size: number;
	total: number;
}

export interface ApiFieldError {
	field: string;
	code: string;
	message: string;
}

/** Единый формат ошибки бэкенда */
export interface ApiError {
	code: string;
	message: string;
	fields?: ApiFieldError[];
	traceId?: string;
}
