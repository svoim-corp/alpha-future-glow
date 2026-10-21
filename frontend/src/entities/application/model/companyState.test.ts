import { companyStateLabel, isCompanyActive } from './companyState';

describe('companyStateLabel', () => {
	it('переводит известные состояния и не теряет неизвестные', () => {
		expect(companyStateLabel('ACTIVE')).toBe('действует');
		expect(companyStateLabel('LIQUIDATING')).toBe('в процессе ликвидации');
		expect(companyStateLabel('SOMETHING')).toBe('SOMETHING');
	});

	it('активна только ACTIVE', () => {
		expect(isCompanyActive('ACTIVE')).toBe(true);
		expect(isCompanyActive('LIQUIDATING')).toBe(false);
	});
});
