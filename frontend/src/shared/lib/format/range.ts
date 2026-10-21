import { formatMoney } from './money';

/** Диапазон платежа: 76300–82300 → «76 300 – 82 300 ₽» */
export const formatMoneyRange = (min: number, max: number): string => {
	if (min === max) {
		return formatMoney(min);
	}

	return `${formatMoney(min).replace(/\s?₽$/u, '')} – ${formatMoney(max)}`;
};
