import { baseApi } from '@shared/api';

export interface AnswerArgs {
	applicationId: string;
	requestId: string;
	answers: { key: string; value: string }[];
}

export const answerApi = baseApi.injectEndpoints({
	endpoints: (build) => ({
		answerInfoRequest: build.mutation<{ requestStatus: string; applicationStatus: string }, AnswerArgs>({
			query: ({ applicationId, requestId, answers }) => ({
				url: `/applications/${applicationId}/info-requests/${requestId}/answer`,
				method: 'POST',
				body: { answers },
			}),
			invalidatesTags: (_result, _error, { applicationId }) => [
				{ type: 'Application', id: applicationId },
				{ type: 'Application', id: 'LIST' },
			],
		}),
	}),
});

export const { useAnswerInfoRequestMutation } = answerApi;
