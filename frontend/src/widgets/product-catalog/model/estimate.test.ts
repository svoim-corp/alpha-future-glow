import { estimateParams } from './estimate';

describe('estimateParams', () => {
	it('собирает параметры расчёта, если введены сумма и срок', () => {
		expect(estimateParams('WORKING', 1500000, 24)).toEqual({ productCode: 'WORKING', amount: 1500000, term: 24 });
	});

	it('без суммы или срока запрос не нужен', () => {
		expect(estimateParams('WORKING', null, 24)).toBeNull();
		expect(estimateParams('WORKING', 1500000, null)).toBeNull();
	});
});
