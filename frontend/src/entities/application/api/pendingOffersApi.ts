import { baseApi, getErrorMessage, type Paginated } from '@shared/api';

import { isPendingOffer, type PendingOffer } from '../model/pendingOffer';
import { type Offer } from '../model/tracking';
import { type ApplicationShort } from '../model/types';

/** У сервера нет отдельного списка офферов: собираем его из всех страниц заявок компании. */
export const pendingOffersApi = baseApi.injectEndpoints({
	endpoints: (build) => ({
		getPendingOffers: build.query<PendingOffer[], void>({
			queryFn: async (_arg, _api, _options, baseQuery) => {
				const applications = new Map<string, ApplicationShort>();
				let page = 1;
				let pages = 1;
				do {
					const result = await baseQuery(`/applications?mine=true&page=${page}&size=100`);
					if (result.error) return { error: result.error };
					const response = result.data as Paginated<ApplicationShort>;
					for (const application of response.items) {
						if (isPendingOffer(application)) applications.set(application.id, application);
					}
					pages = Math.ceil(response.total / response.size);
					page += 1;
				} while (page <= pages);

				const entries = [...applications.values()];
				const data: PendingOffer[] = [];
				// Ограничиваем число одновременных запросов условий.
				for (let index = 0; index < entries.length; index += 5) {
					const batch = await Promise.all(entries.slice(index, index + 5).map(async (application): Promise<PendingOffer> => {
						const result = await baseQuery(`/applications/${application.id}/offer`);
						if (result.error) return { application, offer: null, error: getErrorMessage(result.error, 'Не удалось загрузить условия банка.') };
						const offer = result.data as Offer;
						return offer.decision === 'REJECTED'
							? { application, offer: null, error: 'Решение банка изменилось. Обновите данные или откройте заявку.' }
							: { application, offer };
					}));
					data.push(...batch);
				}
				return { data };
			},
			providesTags: (result) => [
				{ type: 'Application', id: 'LIST' },
				...(result ?? []).map(({ application }) => ({ type: 'Application' as const, id: application.id })),
			],
		}),
	}),
});

export const { useGetPendingOffersQuery } = pendingOffersApi;
