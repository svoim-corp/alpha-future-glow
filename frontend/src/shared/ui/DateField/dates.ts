/** Локальная календарная дата: не используем UTC-преобразования, смещающие выбранный день. */
export const parseCalendarDate = (value: string): Date | undefined => {
	const iso = /^(\d{4})-(\d{2})-(\d{2})$/u.exec(value);
	const local = /^(\d{2})\.(\d{2})\.(\d{4})$/u.exec(value);
	if (!iso && !local) return undefined;
	const [year, month, day] = iso
		? [Number(iso[1]), Number(iso[2]), Number(iso[3])]
		: [Number(local?.[3]), Number(local?.[2]), Number(local?.[1])];
	const date = new Date(0);
	date.setFullYear(year, month - 1, day);
	date.setHours(12, 0, 0, 0);
	return date.getFullYear() === year && date.getMonth() === month - 1 && date.getDate() === day ? date : undefined;
};

export const calendarIso = (date: Date): string => (
	`${String(date.getFullYear()).padStart(4, '0')}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
);

export const displayCalendarDate = (value: string): string => {
	const date = parseCalendarDate(value);
	if (!date) return value;
	return `${String(date.getDate()).padStart(2, '0')}.${String(date.getMonth() + 1).padStart(2, '0')}.${String(date.getFullYear()).padStart(4, '0')}`;
};
