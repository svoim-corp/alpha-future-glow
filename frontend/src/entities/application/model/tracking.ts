import { type ApplicationStatus } from './types';

export type TimelineState = 'DONE' | 'CURRENT' | 'PENDING';
export type Actor = 'CLIENT' | 'BANK';

export interface TimelineStep {
	status: ApplicationStatus;
	label: string;
	at: string | null;
	state: TimelineState;
}

/** GET /applications/{id}/timeline */
export interface Timeline {
	current: {
		status: ApplicationStatus;
		label: string;
		whatsHappening: string;
		nextStep: string;
		actor: Actor;
	};
	steps: TimelineStep[];
}

export interface InfoRequestItem {
	key: string;
	type: 'TEXT';
	label: string;
}

export interface InfoRequest {
	id: string;
	status: 'OPEN' | 'ANSWERED';
	comment: string;
	items: InfoRequestItem[];
	/** Ответы клиента: поле приходит только после ответа */
	answers?: { key: string; value: string }[];
	createdAt: string;
	answeredAt?: string;
}

export interface OfferTerms {
	productCode: string;
	amount: number;
	term: number;
	rate?: number;
	monthlyPayment?: number;
}

export interface OfferApproved {
	decision: 'APPROVED' | 'COUNTER_OFFER';
	requested: OfferTerms;
	offered: Required<OfferTerms>;
	explanation: string | null;
	validUntil: string;
}

export interface OfferRejected {
	decision: 'REJECTED';
	reason: string;
	/** Бэкенд не присылает поле, если альтернатив нет */
	alternatives?: { productCode: string; maxAmount: number }[];
}

/** GET /applications/{id}/offer; до решения сервер отвечает 404 */
export type Offer = OfferApproved | OfferRejected;
