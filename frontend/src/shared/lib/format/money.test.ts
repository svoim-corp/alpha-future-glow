import { formatMoney, formatTerm } from './money';

// Intl использует неразрывные пробелы — нормализуем для читаемости ожиданий
const normalize = (value: string) => value.replace(/\s/gu, ' ');

describe('formatMoney', () => {
	it('форматирует целые рубли с разделителями разрядов', () => {
		expect(normalize(formatMoney(1_500_000))).toBe('1 500 000 ₽');
	});

	it('округляет копейки', () => {
		expect(normalize(formatMoney(999.6))).toBe('1 000 ₽');
	});

	it('форматирует ноль', () => {
		expect(normalize(formatMoney(0))).toBe('0 ₽');
	});
});

describe('formatTerm', () => {
	it.each([
		[1, '1 месяц'],
		[3, '3 месяца'],
		[12, '12 месяцев'],
		[21, '21 месяц'],
		[24, '24 месяца'],
	])('%i → %s', (months, expected) => {
		expect(formatTerm(months)).toBe(expected);
	});
});
