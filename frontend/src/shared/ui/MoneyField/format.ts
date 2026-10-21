/** Форматируем строку, не превращая большие суммы в неточный number. */
export const groupRubles = (digits: string): string => digits.replace(/\B(?=(\d{3})+(?!\d))/gu, ' ');

export const digitsBefore = (text: string, position: number): number => (
	text.slice(0, position).replace(/\D/gu, '').length
);

export const caretAfterDigits = (text: string, count: number): number => {
	if (count === 0) return 0;
	let digits = 0;
	for (let index = 0; index < text.length; index += 1) {
		if (/\d/u.test(text[index] ?? '')) digits += 1;
		if (digits === count) return index + 1;
	}
	return text.length;
};
