import { formatDate, formatDateTime, formatMoneyRange, formatRate, formatRateRange } from './index';

const normalize = (value: string) => value.replace(/\s/gu, ' ');

describe('formatDate', () => {
	it('ISO-дата → дд.мм.гггг', () => {
		expect(formatDate('2026-10-20T12:00:00Z')).toBe('20.10.2026');
	});
});

describe('formatDateTime', () => {
	it('содержит дату и время', () => {
		expect(formatDateTime('2026-10-20T12:00:00Z')).toMatch(/^20\.10\.2026, \d{2}:\d{2}$/u);
	});
});

describe('formatRate', () => {
	it('форматирует ставку с запятой', () => {
		expect(normalize(formatRate(22.5))).toBe('22,5 % годовых');
		expect(normalize(formatRate(22))).toBe('22 % годовых');
	});
});

describe('formatRateRange', () => {
	it('диапазон и одинаковые границы', () => {
		expect(normalize(formatRateRange(20, 28))).toBe('20–28 % годовых');
		expect(normalize(formatRateRange(22, 22))).toBe('22 % годовых');
	});
});

describe('formatMoneyRange', () => {
	it('диапазон платежа с единицей в конце', () => {
		expect(normalize(formatMoneyRange(76300, 82300))).toBe('76 300 – 82 300 ₽');
	});

	it('одинаковые границы — одно значение', () => {
		expect(normalize(formatMoneyRange(1000, 1000))).toBe('1 000 ₽');
	});
});
