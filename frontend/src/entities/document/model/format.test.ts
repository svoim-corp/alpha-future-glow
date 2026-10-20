import { formatFileSize, usageLabel } from './format';

describe('formatFileSize', () => {
	it.each([[245760, '240 КБ'], [1258291, '1,2 МБ'], [100, '1 КБ']])('%i байт → %s', (bytes, text) => {
		expect(formatFileSize(bytes)).toBe(text);
	});
});

describe('usageLabel', () => {
	it('использует подпись договора, если она есть', () => {
		expect(usageLabel({ applications: 0, label: 'Оборотный кредит' })).toBe('Оборотный кредит');
	});

	it.each([[0, 'Ещё не использован'], [1, 'Использован в 1 заявке'], [2, 'Использован в 2 заявках']])(
		'%i использований',
		(applications, text) => {
			expect(usageLabel({ applications, label: null })).toBe(text);
		},
	);
});
