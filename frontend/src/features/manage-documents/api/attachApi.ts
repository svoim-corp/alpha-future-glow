import { baseApi } from '@shared/api';

export const attachDocumentApi = baseApi.injectEndpoints({
	endpoints: (build) => ({
		attachDocument: build.mutation<void, { applicationId: string; documentId: string }>({
			query: ({ applicationId, documentId }) => ({
				url: `/applications/${applicationId}/documents`,
				method: 'POST',
				body: { documentId },
			}),
			invalidatesTags: [{ type: 'Document', id: 'LIST' }],
		}),
	}),
});

export const { useAttachDocumentMutation } = attachDocumentApi;
