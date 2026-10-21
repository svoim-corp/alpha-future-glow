import { calendarIso, displayCalendarDate, parseCalendarDate } from './dates';

describe('календарные даты без смещения часового пояса', () => {
	it('принимает русский и ISO-формат и сохраняет один календарный день', () => {
		const date = parseCalendarDate('29.02.2028');
		expect(date).toBeDefined();
		if (!date) throw new Error('Дата должна быть разобрана');
		expect(calendarIso(date)).toBe('2028-02-29');
		expect(displayCalendarDate('2028-02-29')).toBe('29.02.2028');
	});
	it.each(['31.02.2027', '2027-02-29', '', '02.03.', '2027-13-01'])('отклоняет незавершённую или несуществующую дату %s', (value) => {
		expect(parseCalendarDate(value)).toBeUndefined();
		expect(displayCalendarDate(value)).toBe(value);
	});
});
