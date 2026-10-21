import {
	type BankApplicationDetails,
	type BankQueueItem,
	type BankStats,
	type Completeness,
	type ExternalCheck,
	type HistoryEntry,
	type Score,
} from '@entities/application';

export const bankStatsFixture: BankStats = {
	SUBMITTED: 3,
	DATA_CHECK: 2,
	INFO_REQUESTED: 4,
	SCORING: 1,
	ON_DECISION: 5,
	decidedToday: 7,
};

export const queueItemFixture: BankQueueItem = {
	id: 'a-1024',
	number: 'КЗ-2026-001024',
	company: { inn: '7712345678', name: 'ООО «Северная пекарня»' },
	product: { code: 'WORKING', name: 'Оборотный кредит' },
	amount: 1500000,
	term: 24,
	status: 'ON_DECISION',
	score: 640,
	riskBand: 'MEDIUM',
	decisionMode: 'MANUAL',
	assignee: null,
	submittedAt: '2026-10-20T10:25:00Z',
};

export const queuePageFixture = { items: [queueItemFixture], page: 1, size: 20, total: 1 };

export const bankDetailsFixture: BankApplicationDetails = {
	id: 'a-1024',
	number: 'КЗ-2026-001024',
	status: 'ON_DECISION',
	applicant: { fullName: 'Петрова Анна Сергеевна', authority: 'MCHD', mchdValidUntil: '2027-03-01' },
	company: {
		inn: '7712345678',
		name: 'ООО «Северная пекарня»',
		type: 'LLC',
		regDate: '2015-03-12',
		okved: '10.71',
		taxRegime: 'USN_INCOME_MINUS_EXPENSES',
		actualAddress: 'г. Москва, ул. Складская, 3',
	},
	request: {
		productCode: 'WORKING',
		amount: 1500000,
		term: 24,
		goal: 'WORKING_CAPITAL',
		purposeText: 'Закупка муки и упаковки к сезону',
	},
	turnover: { source: 'BANK_ACCOUNTS', avgMonthlyIncome: 1910000 },
	girbo: { year: 2025, revenue: 21400000, netProfit: 2100000 },
	guarantor: { personId: 'p-1', fullName: 'Смирнов Олег Викторович', inn: '771234567890', bkiConsent: true },
	prelim: { limit: 3000000, prelimRate: 22, amountFits: true, choice: null },
};

export const completenessFixture: Completeness = {
	complete: true,
	issues: [{ severity: 'WARN', code: 'TURNOVER_MISMATCH', text: 'Обороты по выписке выше выручки в ГИР БО на 7%' }],
};

export const checksFixture: ExternalCheck[] = [
	{ source: 'EGRUL', status: 'OK', requestedAt: '2026-10-20T10:26:00Z', summary: 'Действует, 11 лет' },
	{ source: 'COURTS', status: 'WARN', requestedAt: '2026-10-20T10:26:00Z', summary: '1 арбитражное дело, 120 000 ₽' },
];

export const scoreFixture: Score = {
	score: 640,
	riskBand: 'MEDIUM',
	autoDecisionEligible: false,
	autoDecisionReason: 'Пограничный балл и сумма выше лимита без выписки',
	factors: [
		{ code: 'BKI_DISCIPLINE', label: 'Без просрочек', impact: 80 },
		{ code: 'DEBT_LOAD', label: 'Долговая нагрузка 47% от оборота', impact: -70 },
	],
	suggested: { decision: 'COUNTER_OFFER', amount: 1200000, rate: 21.5 },
	modelVersion: 'rules-v1',
};

export const historyFixture: HistoryEntry[] = [
	{
		at: '2026-10-20T10:25:00Z',
		actor: { type: 'CLIENT', name: 'Петрова А. С.' },
		action: 'SUBMIT',
		from: 'DRAFT',
		to: 'SUBMITTED',
	},
	{ at: '2026-10-20T10:26:00Z', actor: { type: 'SYSTEM' }, action: 'CHECKS_DONE', from: 'SUBMITTED', to: 'DATA_CHECK' },
];
