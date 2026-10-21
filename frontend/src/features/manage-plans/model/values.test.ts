import { makePlan } from '@test/fixtures/plans';

import { changedFields, initialValues } from './values';

describe('подготовка полей плана', () => {
	it('новый план создаётся без даты и с уведомлениями в кабинете', () => {
		expect(initialValues()).toMatchObject({ dateMode: 'UNSCHEDULED', startDate: '', endDate: '', inApp: true });
	});

	it('не отправляет неизменённые поля и сохранённый email-флаг', () => {
		const plan = makePlan({ notifications: { inApp: true, email: true } });
		expect(changedFields(initialValues(plan), plan)).toEqual({});
	});

	it('очищает описание строкой, даты — только режимом, и сохраняет остальные данные', () => {
		const plan = makePlan({ note: 'Заметка' });
		expect(changedFields({ ...initialValues(plan), dateMode: 'UNSCHEDULED', note: '' }, plan))
			.toEqual({ dateMode: 'UNSCHEDULED', note: '' });
	});

	it('передаёт новые даты, бюджет, цель и настройки только при изменении', () => {
		const plan = makePlan();
		expect(changedFields({ ...initialValues(plan), title: 'Ремонт', goal: 'REAL_ESTATE', budget: '500000',
			startDate: '2026-11-01', endDate: '2026-11-10', inApp: false }, plan)).toEqual({
			title: 'Ремонт', goal: 'REAL_ESTATE', budgetRub: 500000,
			startDate: '2026-11-01', endDate: '2026-11-10', notifications: { inApp: false },
		});
	});
});
