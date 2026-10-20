import { type BankQueueItem, type BankStats } from '@entities/application';

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
