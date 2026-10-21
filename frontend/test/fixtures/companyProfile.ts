import { type CompanyProfile } from '@entities/application';

export const companyProfileFixture: CompanyProfile = {
	source: ['EGRUL', 'TRANSPARENT_BUSINESS'],
	regAddress: 'г. Москва, ул. Пекарская, 7',
	okved: { code: '10.71', name: 'Производство хлеба' },
	actualAddress: 'г. Москва, ул. Складская, 3',
	head: { personId: 'p-1', fullName: 'Смирнов Олег Викторович' },
	founders: [{ personId: 'p-1', fullName: 'Смирнов Олег Викторович', share: 100 }],
	guarantorCandidates: [{ personId: 'p-1', fullName: 'Смирнов Олег Викторович', roles: ['HEAD', 'FOUNDER'] }],
	taxRegime: 'USN_INCOME_MINUS_EXPENSES',
	turnover: { source: 'BANK_ACCOUNTS', avgMonthlyIncome: 1840000, periodMonths: 12 },
	girbo: { accounted: true },
};
