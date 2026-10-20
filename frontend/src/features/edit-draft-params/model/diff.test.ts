import { fullApplication } from '@test/fixtures/applications';

import { buildParamsPatch, toFormValues } from './diff';

const { params } = fullApplication;

describe('buildParamsPatch', () => {
	it('возвращает undefined, если ничего не менялось', () => {
		expect(buildParamsPatch(toFormValues(params), params)).toBeUndefined();
	});

	it('отправляет только изменённые поля', () => {
		const values = { ...toFormValues(params), amount: '2000000', purposeText: '' };

		expect(buildParamsPatch(values, params)).toEqual({ amount: 2000000, purposeText: null });
	});

	it('пустая отсрочка превращается в null', () => {
		const saved = { ...params, graceMonths: 6 };

		expect(buildParamsPatch({ ...toFormValues(saved), graceMonths: '' }, saved)).toEqual({ graceMonths: null });
	});
});
