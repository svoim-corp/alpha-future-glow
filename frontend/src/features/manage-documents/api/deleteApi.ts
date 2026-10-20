import { baseApi } from '@shared/api';

export const deleteDocumentApi = baseApi.injectEndpoints({
	endpoints: (build) => ({
		deleteDocument: build.mutation<void, string>({
			query: (id) => ({ url: `/documents/${id}`, method: 'DELETE' }),
			invalidatesTags: [{ type: 'Document', id: 'LIST' }],
		}),
	}),
});

export const { useDeleteDocumentMutation } = deleteDocumentApi;
