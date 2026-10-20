import { fullApplication } from '@test/fixtures/applications';

import { buildStage2Patch, toStage2Values } from './diff';

const { stage2 } = fullApplication;

describe('buildStage2Patch', () => {
	it('без изменений ничего не отправляет', () => {
		expect(buildStage2Patch(toStage2Values(stage2), stage2)).toBeUndefined();
	});

	it('фактический адрес отправляется, когда он отличается от регистрации', () => {
		const values = { ...toStage2Values(stage2), sameAsRegistration: false, address: 'г. Москва, ул. Складская, 3' };

		expect(buildStage2Patch(values, stage2)).toEqual({
			actualAddress: { sameAsRegistration: false, address: 'г. Москва, ул. Складская, 3' },
		});
	});

	it('поручитель отправляется только при заполненных ФИО и ИНН', () => {
		const half = { ...toStage2Values(stage2), guarantorInn: '' };
		const full = { ...toStage2Values(stage2), guarantorInn: '771234567891' };

		expect(buildStage2Patch(half, stage2)).toBeUndefined();
		expect(buildStage2Patch(full, stage2)).toEqual({
			guarantor: { fullName: 'Смирнов Олег Викторович', inn: '771234567891' },
		});
	});
});
