import { baseApi } from '@shared/api';

export interface SignResult {
	signed: boolean;
	signedAt: string;
	signatureId: string;
}

export interface DisburseResult {
	status: string;
	loanId: string;
	disbursedAt: string;
	accountMasked: string;
}

interface KeyedArgs {
	applicationId: string;
	idempotencyKey: string;
}

const invalidate = (_result: unknown, _error: unknown, { applicationId }: KeyedArgs) => [
	{ type: 'Application' as const, id: applicationId },
	{ type: 'Application' as const, id: 'LIST' },
	{ type: 'Loan' as const, id: 'LIST' },
];

export const signApi = baseApi.injectEndpoints({
	endpoints: (build) => ({
		signContract: build.mutation<SignResult, KeyedArgs>({
			query: ({ applicationId, idempotencyKey }) => ({
				url: `/applications/${applicationId}/contract/sign`,
				method: 'POST',
				body: { confirm: true },
				headers: { 'Idempotency-Key': idempotencyKey },
			}),
			invalidatesTags: invalidate,
		}),
		disburseLoan: build.mutation<DisburseResult, KeyedArgs>({
			query: ({ applicationId, idempotencyKey }) => ({
				url: `/applications/${applicationId}/contract/disburse`,
				method: 'POST',
				headers: { 'Idempotency-Key': idempotencyKey },
			}),
			invalidatesTags: invalidate,
		}),
	}),
});

export const { useSignContractMutation, useDisburseLoanMutation } = signApi;
