import { baseApi, type Paginated } from '@shared/api';

import {
	type Application,
	type ApplicationShort,
	type CompanyProfile,
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
		getCompanyProfile: build.query<CompanyProfile, string>({
			query: (id) => `/applications/${id}/company-profile`,
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
			onQueryStarted: async ({ id, body }, { dispatch, queryFulfilled }) => {
				try {
					const { data } = await queryFulfilled;
					dispatch(applicationApi.util.updateQueryData('getApplication', id, (draft) => {
						draft.version = data.version;
						draft.savedAt = data.savedAt;
					}));

					if (data.prelimInvalidated || body.stage2?.guarantor) {
						// Перечитываем состояние этапов, поручителя и его согласия, вычисляемое сервером.
						dispatch(applicationApi.util.invalidateTags([{ type: 'Application', id }]));
					}
				} catch {
					// ошибку показывает вызывающий код через unwrap()
				}
			},
		}),
	}),
});

export const {
	useGetMyApplicationsQuery,
	useGetApplicationQuery,
	useGetCompanyProfileQuery,
	usePatchApplicationMutation,
} = applicationApi;
