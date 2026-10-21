import { type Application, type ApplicationShort } from '@entities/application';

export const infoRequestedApplication: ApplicationShort = {
	id: 'a-1024',
	number: 'КЗ-2026-001024',
	product: { code: 'WORKING', name: 'Оборотный кредит' },
	amount: 1500000,
	term: 24,
	status: 'INFO_REQUESTED',
	statusLabel: 'Нужна доп. информация',
	actionRequired: true,
	updatedAt: '2026-10-20T12:40:00Z',
};

export const draftApplication: ApplicationShort = {
	id: 'a-1031',
	number: null,
	product: { code: 'EXPRESS', name: 'Экспресс-кредит' },
	amount: 500000,
	term: 12,
	status: 'DRAFT',
	statusLabel: 'Черновик',
	actionRequired: false,
	updatedAt: '2026-10-19T09:10:00Z',
};

/** Полная заявка на этапе 1 с готовым согласием и лимитом — как в примере API_SPEC, 4.5 */
export const fullApplication: Application = {
	id: 'a-1031',
	number: null,
	status: 'DRAFT',
	stage: 2,
	version: 7,
	product: { code: 'WORKING', name: 'Оборотный кредит' },
	params: {
		amount: 1500000,
		term: 24,
		goal: 'WORKING_CAPITAL',
		purposeText: 'Закупка муки и упаковки к сезону',
		graceMonths: null,
		overdraftLimit: null,
	},
	company: {
		inn: '7712345678',
		type: 'LLC',
		name: 'ООО «Северная пекарня»',
		ogrn: '1157746000000',
		regDate: '2015-03-12',
		state: 'ACTIVE',
	},
	authority: { type: 'HEAD', verified: true },
	consents: { bki: { given: true, at: '2026-10-20T10:05:00Z' } },
	prelim: {
		outcome: 'LIMIT',
		limit: 3000000,
		prelimRate: 22,
		amountFits: true,
		choice: null,
	},
	stage2: {
		actualAddress: 'г. Москва, ул. Складская, 3',
		turnover: { source: 'BANK_ACCOUNTS', avgMonthlyIncome: 1840000, periodMonths: 12 },
		guarantor: { personId: 'p-1', fullName: 'Смирнов Олег Викторович', bkiConsent: true },
		truthConfirmed: false,
	},
	locks: { paramsLocked: false, stage2Available: true },
	savedAt: '2026-10-20T10:12:00Z',
};
