/** Целое положительное число из текстового поля; пробелы-разделители допустимы. Иначе null */
export const parseWholeNumber = (text: string): number | null => {
	const digits = text.replace(/\s/gu, '');

	if (!/^\d+$/u.test(digits)) {
		return null;
	}

	const value = Number(digits);

	return value > 0 && Number.isSafeInteger(value) ? value : null;
};

export const MAX_TERM_MONTHS = 120;

export interface CatalogInputErrors {
	amount?: string;
	term?: string;
}

/** Проверка полей подбора: сообщения показываем под полями */
export const validateCatalogInput = (amountText: string, termText: string): CatalogInputErrors => {
	const errors: CatalogInputErrors = {};
	const term = parseWholeNumber(termText);

	if (parseWholeNumber(amountText) === null) {
		errors.amount = 'Введите сумму в рублях целым числом';
	}

	if (term === null || term > MAX_TERM_MONTHS) {
		errors.term = `Введите срок в месяцах от 1 до ${MAX_TERM_MONTHS}`;
	}

	return errors;
};
