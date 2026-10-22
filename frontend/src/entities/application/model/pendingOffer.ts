import { type OfferApproved } from './tracking';
import { type ApplicationShort } from './types';

export interface PendingOffer {
	application: ApplicationShort;
	offer: OfferApproved | null;
	error?: string;
}

export const isPendingOffer = (application: ApplicationShort) =>
	['APPROVED', 'COUNTER_OFFER', 'SIGNING'].includes(application.status);

export const isExpiredOffer = (entry: PendingOffer, today: string) =>
	entry.application.status !== 'SIGNING' && entry.offer !== null && entry.offer.validUntil < today;
