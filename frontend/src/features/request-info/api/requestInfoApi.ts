import { baseApi } from '@shared/api';

export interface RequestInfoArgs {
	applicationId: string;
	comment: string;
	items: { key: string; type: 'TEXT'; label: string }[];
}

export const requestInfoApi = baseApi.injectEndpoints({
	endpoints: (build) => ({
		requestInfo: build.mutation<{ id: string; status: string; applicationStatus: string }, RequestInfoArgs>({
			query: ({ applicationId, ...body }) => ({
				url: `/bank/applications/${applicationId}/info-requests`,
				method: 'POST',
				body,
			}),
			invalidatesTags: (_result, _error, { applicationId }) => [
				{ type: 'BankApplication', id: applicationId },
				{ type: 'BankApplication', id: 'LIST' },
			],
		}),
	}),
});

export const { useRequestInfoMutation } = requestInfoApi;
