import { planSchema, type PlanValues } from './schema';

const base: PlanValues = {
	title: 'Новая печь',
	budget: '1200000',
	goal: 'EQUIPMENT',
	dateMode: 'UNSCHEDULED',
	startDate: '',
	endDate: '',
	note: '',
	inApp: true,
};
const messages = (values: PlanValues) => {
	const result = planSchema.safeParse(values);

	return result.success ? [] : result.error.issues.map((issue) => issue.message);
};

describe('planSchema', () => {
	it('идея без даты не требует дат', () => {
		expect(messages(base)).toEqual([]);
	});

	it('для дня и периода нужна дата начала', () => {
		expect(messages({ ...base, dateMode: 'DAY' })).toEqual(['Укажите дату начала']);
	});

	it('период требует окончания не раньше начала', () => {
		expect(messages({ ...base, dateMode: 'PERIOD', startDate: '2027-06-01' })).toEqual(['Укажите дату окончания']);
		expect(messages({ ...base, dateMode: 'PERIOD', startDate: '2027-06-10', endDate: '2027-06-01' }))
			.toEqual(['Окончание не может быть раньше начала']);
		expect(messages({ ...base, dateMode: 'PERIOD', startDate: '2027-06-01', endDate: '2027-06-01' })).toEqual([]);
	});

	it('цель обязательна, ограничения длины соответствуют серверу', () => {
		expect(messages({ ...base, goal: '' })).toContain('Опишите цель');
		expect(messages({ ...base, goal: '   ' })).toContain('Опишите цель');
		expect(messages({ ...base, goal: 'я'.repeat(65) })).toContain('Не больше 64 символов');
		expect(planSchema.parse({ ...base, goal: '  Своя задача  ' }).goal).toBe('Своя задача');
		expect(messages({ ...base, title: 'Я', note: 'а'.repeat(2000) })).toEqual([]);
		expect(messages({ ...base, title: 'а'.repeat(201) })).toContain('Не больше 200 символов');
		expect(messages({ ...base, note: 'а'.repeat(2001) })).toContain('Не больше 2000 символов');
	});

	it('не принимает несуществующую дату, но принимает високосный день', () => {
		expect(messages({ ...base, dateMode: 'DAY', startDate: '2027-02-29' })).toContain('Укажите корректную дату');
		expect(messages({ ...base, dateMode: 'DAY', startDate: '2028-02-29' })).toEqual([]);
	});

	it('не допускает потерю точности бюджета', () => {
		expect(messages({ ...base, budget: '9007199254740992' })).toContain('Сумма слишком велика');
	});

	it('бюджет — положительное целое', () => {
		expect(messages({ ...base, budget: '0' })).toEqual(['Бюджет должен быть больше нуля']);
		expect(messages({ ...base, budget: '12,5' })).toEqual(['Бюджет — целое число рублей']);
	});
});
