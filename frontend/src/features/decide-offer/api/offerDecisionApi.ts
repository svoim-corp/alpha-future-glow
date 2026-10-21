import { baseApi } from '@shared/api';

const invalidate = (_result: unknown, _error: unknown, { applicationId }: { applicationId: string }) => [
	{ type: 'Application' as const, id: applicationId },
	{ type: 'Application' as const, id: 'LIST' },
];

export const offerDecisionApi = baseApi.injectEndpoints({
	endpoints: (build) => ({
		acceptOffer: build.mutation<{ status: string; redirect: string }, { applicationId: string }>({
			query: ({ applicationId }) => ({ url: `/applications/${applicationId}/offer/accept`, method: 'POST' }),
			invalidatesTags: invalidate,
		}),
		declineOffer: build.mutation<{ status: string }, { applicationId: string; reason: string }>({
			query: ({ applicationId, reason }) => ({
				url: `/applications/${applicationId}/offer/decline`,
				method: 'POST',
				body: { reason },
			}),
			invalidatesTags: invalidate,
		}),
	}),
});

export const { useAcceptOfferMutation, useDeclineOfferMutation } = offerDecisionApi;
