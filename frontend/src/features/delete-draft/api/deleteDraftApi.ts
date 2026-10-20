import { baseApi } from '@shared/api';

export const deleteDraftApi = baseApi.injectEndpoints({
	endpoints: (build) => ({
		deleteDraft: build.mutation<void, string>({
			query: (id) => ({ url: `/applications/${id}`, method: 'DELETE' }),
			invalidatesTags: (_result, _error, id) => [{ type: 'Application', id: 'LIST' }, { type: 'Application', id }],
		}),
	}),
});

export const { useDeleteDraftMutation } = deleteDraftApi;
