const pad = (value: number): string => String(value).padStart(2, '0');

/** Дата без часового пояса: календарь работает с днями, а не с моментами времени */
export const toIsoDate = (year: number, month: number, day: number): string => `${year}-${pad(month + 1)}-${pad(day)}`;

export interface MonthRef {
	year: number;
	/** 0–11 */
	month: number;
}

export const shiftMonth = ({ year, month }: MonthRef, delta: number): MonthRef => {
	const index = year * 12 + month + delta;

	return { year: Math.floor(index / 12), month: ((index % 12) + 12) % 12 };
};

export const monthRange = ({ year, month }: MonthRef): { from: string; to: string } => ({
	from: toIsoDate(year, month, 1),
	to: toIsoDate(year, month, new Date(year, month + 1, 0).getDate()),
});

/** Недели месяца, начиная с понедельника; дни вне месяца — null */
export const buildMonthGrid = ({ year, month }: MonthRef): (string | null)[][] => {
	const daysInMonth = new Date(year, month + 1, 0).getDate();
	const offset = (new Date(year, month, 1).getDay() + 6) % 7;
	const cells: (string | null)[] = [
		...Array.from({ length: offset }, () => null),
		...Array.from({ length: daysInMonth }, (_, index) => toIsoDate(year, month, index + 1)),
	];

	while (cells.length % 7 !== 0) {
		cells.push(null);
	}

	return Array.from({ length: cells.length / 7 }, (_, week) => cells.slice(week * 7, week * 7 + 7));
};

export interface MonthCell {
	date: string;
	/** День соседнего месяца, дорисованный до полной недели */
	outside: boolean;
}

/** Недели месяца с днями соседних месяцев вместо пустых клеток — как в макете планера */
export const buildMonthCells = ({ year, month }: MonthRef): MonthCell[][] => {
	const offset = (new Date(year, month, 1).getDay() + 6) % 7;
	const daysInMonth = new Date(year, month + 1, 0).getDate();
	const weeks = Math.ceil((offset + daysInMonth) / 7);

	return Array.from({ length: weeks }, (_, week) => Array.from({ length: 7 }, (_, day) => {
		const date = new Date(year, month, 1 - offset + week * 7 + day);

		return {
			date: toIsoDate(date.getFullYear(), date.getMonth(), date.getDate()),
			outside: date.getMonth() !== month,
		};
	}));
};
