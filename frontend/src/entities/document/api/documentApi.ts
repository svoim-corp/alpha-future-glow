import { baseApi } from '@shared/api';

import { type DocumentFilters, type DocumentsResponse } from '../model/types';

const toQuery = ({ category, q, sort }: DocumentFilters): string => {
	const params = new URLSearchParams({ sort });

	if (category) {
		params.set('category', category);
	}

	if (q?.trim()) {
		params.set('q', q.trim());
	}

	return params.toString();
};

export const documentApi = baseApi.injectEndpoints({
	endpoints: (build) => ({
		getDocuments: build.query<DocumentsResponse, DocumentFilters>({
			query: (filters) => `/documents?${toQuery(filters)}`,
			providesTags: [{ type: 'Document', id: 'LIST' }],
		}),
	}),
});

export const { useGetDocumentsQuery } = documentApi;
