import { baseApi } from '@shared/api';

export type CheckStatus = 'OK' | 'WARN' | 'FAIL';

export interface PrelimCheck {
	source: string;
	status: CheckStatus;
	label: string;
}

export interface PrelimLimit {
	outcome: 'LIMIT';
	requestedAmount: number;
	limit: number;
	prelimRate: number;
	prelimPayment: number;
	/** false — запрошенная сумма больше лимита */
	amountFits: boolean;
	checks: PrelimCheck[];
}

export interface PrelimStop {
	outcome: 'STOP';
	stopReason: string;
	stopText: string;
	whatToDo: string[];
}

export type PrelimResult = PrelimLimit | PrelimStop;

export interface ChoiceArgs {
	applicationId: string;
	choice: 'REDUCE_AMOUNT' | 'KEEP_AMOUNT';
	/** Для REDUCE_AMOUNT — новая сумма, не больше лимита */
	amount?: number;
}

export interface ChoiceResult {
	stage: 1 | 2;
	params: { amount: number; term: number };
	version: number;
}

export const BKI_CONSENT_TEXT = 'Согласен на запрос моей кредитной истории в бюро кредитных историй';

export const prelimApi = baseApi.injectEndpoints({
	endpoints: (build) => ({
		giveBkiConsent: build.mutation<{ given: boolean; at: string }, string>({
			query: (applicationId) => ({
				url: `/applications/${applicationId}/consents/bki`,
				method: 'POST',
				body: { given: true, text: BKI_CONSENT_TEXT },
			}),
			invalidatesTags: (_result, _error, id) => [{ type: 'Application', id }],
		}),
		requestPrelim: build.mutation<PrelimResult, string>({
			query: (applicationId) => ({ url: `/applications/${applicationId}/prelim`, method: 'POST' }),
			invalidatesTags: (_result, _error, id) => [{ type: 'Application', id }],
		}),
		choosePrelim: build.mutation<ChoiceResult, ChoiceArgs>({
			query: ({ applicationId, ...body }) => ({
				url: `/applications/${applicationId}/prelim/choice`,
				method: 'POST',
				body,
			}),
			invalidatesTags: (_result, _error, { applicationId }) => [{ type: 'Application', id: applicationId }],
		}),
	}),
});

export const { useGiveBkiConsentMutation, useRequestPrelimMutation, useChoosePrelimMutation } = prelimApi;
