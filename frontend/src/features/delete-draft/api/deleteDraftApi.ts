import { baseApi } from '@shared/api';

export const deleteDraftApi = baseApi.injectEndpoints({
	endpoints: (build) => ({
		deleteDraft: build.mutation<void, string>({
			query: (id) => ({ url: `/applications/${encodeURIComponent(id)}`, method: 'DELETE' }),
			invalidatesTags: (_result, error, id) => error ? [] : [
				{ type: 'Application', id: 'LIST' }, { type: 'Application', id }, 'SupportChat',
			],
		}),
	}),
});

export const { useDeleteDraftMutation } = deleteDraftApi;
