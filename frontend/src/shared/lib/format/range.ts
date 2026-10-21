import { formatMoney } from './money';

/** Диапазон платежа: 76300–82300 → «76 300 – 82 300 ₽» */
export const formatMoneyRange = (min: number, max: number): string => {
	if (min === max) {
		return formatMoney(min);
	}

	return `${formatMoney(min).replace(/\s?₽$/u, '')} – ${formatMoney(max)}`;
};

const compactFormatter = new Intl.NumberFormat('ru-RU', { maximumFractionDigits: 1 });

const compactParts = (value: number): { number: string; unit: string } => {
	if (value >= 1000000) {
		return { number: compactFormatter.format(value / 1000000), unit: 'млн' };
	}

	return value >= 1000 ? { number: compactFormatter.format(value / 1000), unit: 'тыс.' } : { number: String(value), unit: '' };
};

/** Диапазон суммы для карточки: «300 тыс. — 10 млн ₽», «1–30 млн ₽» */
export const formatCompactMoneyRange = (min: number, max: number): string => {
	const from = compactParts(min);
	const to = compactParts(max);

	if (from.unit === to.unit) {
		return `${from.number}–${to.number} ${to.unit} ₽`.replace('  ', ' ');
	}

	return `${from.number} ${from.unit} — ${to.number} ${to.unit} ₽`.replace(/\s+/gu, ' ');
};

/** Одна граница суммы: 3000000 → «3 млн ₽» */
export const formatCompactMoney = (value: number): string => {
	const part = compactParts(value);

	return `${part.number} ${part.unit} ₽`.replace(/\s+/gu, ' ');
};
