import { type Range } from '@entities/product';

export const clamp = (value: number, range: Range) => Math.min(range.max, Math.max(range.min, value));

/** Доля заполнения ползунка в процентах — для цвета трека */
export const fillPercent = (value: number, range: Range) => (
	range.max === range.min ? 0 : ((value - range.min) / (range.max - range.min)) * 100
);

export const DEFAULT_AMOUNT = 1500000;
export const DEFAULT_TERM = 24;
