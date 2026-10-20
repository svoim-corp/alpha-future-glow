import { parseWholeNumber, validateCatalogInput } from './input';

describe('parseWholeNumber', () => {
	it.each([
		['1500000', 1500000],
		['1 500 000', 1500000],
		['1 500 000', 1500000],
		['24', 24],
	])('«%s» → %i', (text, expected) => {
		expect(parseWholeNumber(text)).toBe(expected);
	});

	it.each([[''], ['abc'], ['0'], ['-5'], ['1,5'], ['12abc'], ['99999999999999999999']])('«%s» → null', (text) => {
		expect(parseWholeNumber(text)).toBeNull();
	});
});

describe('validateCatalogInput', () => {
	it('валидные значения — без ошибок', () => {
		expect(validateCatalogInput('1500000', '24')).toEqual({});
	});

	it('пустая сумма и срок вне диапазона дают ошибки у своих полей', () => {
		expect(validateCatalogInput('', '24').amount).toBe('Введите сумму в рублях целым числом');
		expect(validateCatalogInput('1000', '121').term).toBe('Введите срок в месяцах от 1 до 120');
		expect(validateCatalogInput('1000', '0').term).toBeDefined();
	});
});
