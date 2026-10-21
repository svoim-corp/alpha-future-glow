import { investDetails, overdraftDetails, workingDetails } from '@test/fixtures/products';

import { buildConditionRows } from './conditions';

const goals = [
	{ code: 'WORKING_CAPITAL', name: 'Пополнить оборотные средства' },
	{ code: 'EQUIPMENT', name: 'Купить оборудование' },
];

type Rows = ReturnType<typeof buildConditionRows>;
const asMap = (rows: Rows) => Object.fromEntries(rows.map((row) => [row.label, row.value]));
const normalize = (value: unknown) => String(value).replace(/\s/gu, ' ');

describe('buildConditionRows', () => {
	it('оборотный кредит: цели по названиям, диапазоны', () => {
		const rows = asMap(buildConditionRows(workingDetails, goals));

		expect(rows['Цели']).toBe('Пополнить оборотные средства');
		expect(normalize(rows['Сумма'])).toBe('300 000 ₽ – 10 000 000 ₽');
		expect(rows['Срок']).toBe('3–36 мес.');
		expect(normalize(rows['Ставка'])).toBe('20–28 % годовых');
		expect(rows['Отсрочка']).toBeUndefined();
	});

	it('инвестиционный: отсрочка, требования к возрасту и поручитель для ООО, неизвестная цель остаётся кодом', () => {
		const rows = asMap(buildConditionRows(investDetails, goals));

		expect(rows['Отсрочка']).toBe('По основному долгу до 6 мес.');
		expect(rows['Требования']).toBe('бизнесу не меньше 12 мес.; для ООО нужен поручитель');
		expect(rows['Цели']).toBe('Купить оборудование, REAL_ESTATE, EXPANSION');
		expect(rows['Лимит']).toBe('До 1,5 среднемесячных поступлений на счета в банке');
	});

	it('овердрафт: вместо срока и суммы — лимит', () => {
		const rows = asMap(buildConditionRows(overdraftDetails, goals));

		expect(rows['Срок']).toBeUndefined();
		expect(rows['Сумма']).toBeUndefined();
		expect(normalize(rows['Лимит'])).toBe('100 000 ₽ – 5 000 000 ₽');
		expect(rows['Требования']).toBe('Особых требований нет');
	});
});
