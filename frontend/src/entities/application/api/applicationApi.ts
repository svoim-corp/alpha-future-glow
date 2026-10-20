import { baseApi, type Paginated } from '@shared/api';

import {
	type Application,
	type ApplicationShort,
	type PatchApplicationArgs,
	type PatchApplicationResult,
} from '../model/types';

export const applicationApi = baseApi.injectEndpoints({
	endpoints: (build) => ({
		getMyApplications: build.query<Paginated<ApplicationShort>, void>({
			query: () => '/applications?mine=true',
			providesTags: [{ type: 'Application', id: 'LIST' }],
		}),
		getApplication: build.query<Application, string>({
			query: (id) => `/applications/${id}`,
			providesTags: (_result, _error, id) => [{ type: 'Application', id }],
		}),
		/** Автосохранение черновика. Версию передаём в If-Match, ответ несёт новую версию */
		patchApplication: build.mutation<PatchApplicationResult, PatchApplicationArgs>({
			query: ({ id, version, body }) => ({
				url: `/applications/${id}`,
				method: 'PATCH',
				body,
				headers: { 'If-Match': String(version) },
			}),
			onQueryStarted: async ({ id }, { dispatch, queryFulfilled }) => {
				try {
					const { data } = await queryFulfilled;
					dispatch(applicationApi.util.updateQueryData('getApplication', id, (draft) => {
						draft.version = data.version;
						draft.savedAt = data.savedAt;
					}));

					if (data.prelimInvalidated) {
						// сумма, срок или продукт изменились — результат этапа 1 устарел, берём состояние с сервера
						dispatch(applicationApi.util.invalidateTags([{ type: 'Application', id }]));
					}
				} catch {
					// ошибку показывает вызывающий код через unwrap()
				}
			},
		}),
	}),
});

export const { useGetMyApplicationsQuery, useGetApplicationQuery, usePatchApplicationMutation } = applicationApi;
