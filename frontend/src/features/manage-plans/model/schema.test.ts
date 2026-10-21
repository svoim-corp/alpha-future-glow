import { planSchema, type PlanValues } from './schema';

const base: PlanValues = {
	title: 'Новая печь',
	budget: '1200000',
	goal: 'EQUIPMENT',
	dateMode: 'UNSCHEDULED',
	startDate: '',
	endDate: '',
	note: '',
	email: false,
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

	it('бюджет — положительное целое', () => {
		expect(messages({ ...base, budget: '0' })).toEqual(['Бюджет должен быть больше нуля']);
		expect(messages({ ...base, budget: '12,5' })).toEqual(['Бюджет — целое число рублей']);
	});
});
