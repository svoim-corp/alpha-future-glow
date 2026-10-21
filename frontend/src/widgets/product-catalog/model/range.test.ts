import { clamp, fillPercent } from './range';

describe('range', () => {
	it('ограничивает значение рамками продукта', () => {
		expect(clamp(50, { min: 100, max: 200 })).toBe(100);
		expect(clamp(500, { min: 100, max: 200 })).toBe(200);
		expect(clamp(150, { min: 100, max: 200 })).toBe(150);
	});

	it('считает долю заполнения ползунка', () => {
		expect(fillPercent(150, { min: 100, max: 200 })).toBe(50);
		expect(fillPercent(1, { min: 1, max: 1 })).toBe(0);
	});
});
