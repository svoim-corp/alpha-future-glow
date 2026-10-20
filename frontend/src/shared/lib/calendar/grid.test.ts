import { buildMonthGrid, monthRange, shiftMonth, toIsoDate } from './grid';

describe('календарная сетка', () => {
	it('ноябрь 2026 начинается с воскресенья: перед 1-м числом шесть пустых клеток', () => {
		const [first] = buildMonthGrid({ year: 2026, month: 10 });

		expect(first?.slice(0, 6)).toEqual([null, null, null, null, null, null]);
		expect(first?.[6]).toBe('2026-11-01');
	});

	it('каждая неделя содержит семь клеток, а дней в месяце столько, сколько нужно', () => {
		const grid = buildMonthGrid({ year: 2026, month: 1 });

		expect(grid.every((week) => week.length === 7)).toBe(true);
		expect(grid.flat().filter(Boolean)).toHaveLength(28);
	});

	it('границы месяца для запроса расписания', () => {
		expect(monthRange({ year: 2026, month: 11 })).toEqual({ from: '2026-12-01', to: '2026-12-31' });
	});

	it('переход через границу года', () => {
		expect(shiftMonth({ year: 2026, month: 11 }, 1)).toEqual({ year: 2027, month: 0 });
		expect(shiftMonth({ year: 2026, month: 0 }, -1)).toEqual({ year: 2025, month: 11 });
	});

	it('формирует ISO-дату с ведущими нулями', () => {
		expect(toIsoDate(2026, 0, 5)).toBe('2026-01-05');
	});
});
