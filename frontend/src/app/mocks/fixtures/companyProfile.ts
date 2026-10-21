import { type CompanyProfile } from '@entities/application';

export const companyProfileFixture: CompanyProfile = {
	source: ['EGRUL', 'TRANSPARENT_BUSINESS'],
	regAddress: 'г. Москва, ул. Пекарская, 7',
	okved: { code: '10.71', name: 'Производство хлеба' },
	head: 'Смирнов Олег Викторович',
	founders: [{ name: 'Смирнов Олег Викторович', share: 100 }],
	taxRegime: 'USN_INCOME_MINUS_EXPENSES',
	isBankClient: true,
	internalTurnover: { avgMonthlyIncome: 1840000, periodMonths: 12 },
	girbo: { accounted: true },
};
