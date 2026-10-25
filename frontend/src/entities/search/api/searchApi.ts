import { baseApi } from '@shared/api';

import { type SearchArgs, type SearchResponse, searchResponseSchema } from '../model/schema';

export const searchApi = baseApi.injectEndpoints({
	endpoints: (build) => ({
		globalSearch: build.query<SearchResponse, SearchArgs>({
			async queryFn({ query, type, limit, cursor }, _api, _options, baseQuery) {
				const result = await baseQuery({
					url: '/search', cache: 'no-store',
					params: { q: query, types: type, limitPerType: limit, cursor },
				});
				if (result.error) return { error: result.error };
				const parsed = searchResponseSchema.safeParse(result.data);
				if (!parsed.success) return {
					error: { status: 'CUSTOM_ERROR', error: 'Некорректный ответ поиска',
						data: { code: 'SEARCH_RESPONSE_INVALID', message: 'Не удалось прочитать результаты поиска.' } },
				};
				return { data: parsed.data };
			},
			providesTags: [
				{ type: 'Application', id: 'LIST' }, 'Loan', { type: 'Plan', id: 'LIST' },
				'Product', { type: 'Ticket', id: 'LIST' },
			],
			keepUnusedDataFor: 15,
		}),
	}),
});

export const { useGlobalSearchQuery } = searchApi;
