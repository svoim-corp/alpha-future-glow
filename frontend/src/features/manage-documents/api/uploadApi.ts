import { baseApi } from '@shared/api';

export const uploadDocumentApi = baseApi.injectEndpoints({
	endpoints: (build) => ({
		uploadDocument: build.mutation<unknown, { file: File; category: string }>({
			query: ({ file, category }) => {
				const body = new FormData();
				body.append('file', file);
				body.append('category', category);

				return { url: '/documents', method: 'POST', body };
			},
			invalidatesTags: [{ type: 'Document', id: 'LIST' }],
		}),
	}),
});

export const { useUploadDocumentMutation } = uploadDocumentApi;
