import { type InfoRequest, type Offer, type Timeline } from '@entities/application';

export const timelineFixture: Timeline = {
	current: {
		status: 'INFO_REQUESTED',
		label: 'Нужна доп. информация',
		whatsHappening: 'Банк проверил заявку и просит уточнить данные',
		nextStep: 'Ответьте на запрос банка ниже',
		actor: 'CLIENT',
	},
	steps: [
		{ status: 'SUBMITTED', label: 'Подана', at: '2026-10-20T10:25:00Z', state: 'DONE' },
		{ status: 'DATA_CHECK', label: 'Проверка данных', at: '2026-10-20T10:26:00Z', state: 'DONE' },
		{ status: 'INFO_REQUESTED', label: 'Нужна доп. информация', at: '2026-10-20T12:40:00Z', state: 'CURRENT' },
		{ status: 'SCORING', label: 'Обогащение и скоринг', at: null, state: 'PENDING' },
	],
};

export const openInfoRequest: InfoRequest = {
	id: 'r-7',
	status: 'OPEN',
	comment: 'Обороты по счетам расходятся с отчётностью',
	items: [
		{ key: 'actual_address', type: 'TEXT', label: 'Адрес склада, где хранится товар' },
		{ key: 'turnover_note', type: 'TEXT', label: 'Почему выручка за 2025 год ниже оборотов по счёту' },
	],
	createdAt: '2026-10-20T12:40:00Z',
};

export const infoRequestsFixture: InfoRequest[] = [openInfoRequest];

export const counterOfferFixture: Offer = {
	decision: 'COUNTER_OFFER',
	requested: { productCode: 'WORKING', amount: 1500000, term: 24 },
	offered: { productCode: 'WORKING', amount: 1200000, term: 24, rate: 21.5, monthlyPayment: 62200 },
	explanation: 'Сумма снижена по уровню долговой нагрузки',
	validUntil: '2026-10-27',
};

export const rejectedOfferFixture: Offer = {
	decision: 'REJECTED',
	reason: 'Высокая долговая нагрузка по данным БКИ',
	alternatives: [{ productCode: 'EXPRESS', maxAmount: 500000 }],
};
