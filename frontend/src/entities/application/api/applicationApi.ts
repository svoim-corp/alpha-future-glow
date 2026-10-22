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
		/** Полный список для сводок; частичная загрузка не выдаётся за полный результат. */
		getAllMyApplications: build.query<ApplicationShort[], void>({
			queryFn: async (_arg, _api, _options, baseQuery) => {
				const items = new Map<string, ApplicationShort>();
				let expectedTotal: number | undefined;
				for (let page = 1; page <= 1000; page += 1) {
					const result = await baseQuery(`/applications?mine=true&page=${page}&size=100`);
					if (result.error) return { error: result.error };
					const data = result.data as Paginated<ApplicationShort> | null;
					if (!data || !Array.isArray(data.items) || data.page !== page
						|| !Number.isInteger(data.total) || data.total < 0
						|| !Number.isInteger(data.size) || data.size < 1 || data.size > 100
						|| data.items.length > data.size
						|| (expectedTotal !== undefined && expectedTotal !== data.total)) {
						return { error: { status: 'CUSTOM_ERROR', error: 'Сервер вернул некорректную страницу заявок.' } };
					}
					expectedTotal = data.total;
					for (const item of data.items) items.set(item.id, item);
					if (page * data.size >= data.total) {
						if (items.size === data.total) return { data: [...items.values()].sort((a, b) => b.updatedAt.localeCompare(a.updatedAt)) };
						break;
					}
					if (data.items.length === 0) break;
				}
				return { error: { status: 'CUSTOM_ERROR', error: 'Не удалось загрузить полный список заявок.' } };
			},
			providesTags: (result) => [{ type: 'Application', id: 'LIST' }, ...(result ?? []).map((item) => ({ type: 'Application' as const, id: item.id }))],
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
	useGetAllMyApplicationsQuery,
	useGetApplicationQuery,
	useGetCompanyProfileQuery,
	usePatchApplicationMutation,
} = applicationApi;
