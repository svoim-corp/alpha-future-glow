import { baseApi } from '@shared/api';

export interface DecisionArgs {
	applicationId: string;
	idempotencyKey: string;
	body: {
		type: 'APPROVE' | 'COUNTER' | 'REJECT';
		amount?: number;
		term?: number;
		productCode?: string;
		rate?: number;
		reason?: string;
	};
}

export const decisionApi = baseApi.injectEndpoints({
	endpoints: (build) => ({
		makeDecision: build.mutation<{ status: string; decisionId: string; mode: string }, DecisionArgs>({
			query: ({ applicationId, idempotencyKey, body }) => ({
				url: `/bank/applications/${applicationId}/decision`,
				method: 'POST',
				body,
				headers: { 'Idempotency-Key': idempotencyKey },
			}),
			invalidatesTags: (_result, _error, { applicationId }) => [
				{ type: 'BankApplication', id: applicationId },
				{ type: 'BankApplication', id: 'LIST' },
			],
		}),
	}),
});

export const { useMakeDecisionMutation } = decisionApi;
