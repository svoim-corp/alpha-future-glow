const rateFormatter = new Intl.NumberFormat('ru-RU', { maximumFractionDigits: 1 });

/** Ставка в процентах годовых: 22.5 → «22,5 % годовых» */
export const formatRate = (rate: number): string => `${rateFormatter.format(rate)} % годовых`;

/** Диапазон ставки: 20–28 → «20–28 % годовых» */
export const formatRateRange = (min: number, max: number): string => (
	min === max ? formatRate(min) : `${rateFormatter.format(min)}–${rateFormatter.format(max)} % годовых`
);
