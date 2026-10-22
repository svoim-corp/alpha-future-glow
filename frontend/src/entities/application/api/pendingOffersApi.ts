import { baseApi, getErrorMessage } from '@shared/api';

import { isPendingOffer, type PendingOffer } from '../model/pendingOffer';
import { type Offer } from '../model/tracking';
import { applicationApi } from './applicationApi';

/** У сервера нет отдельного списка офферов: собираем его из всех страниц заявок компании. */
export const pendingOffersApi = baseApi.injectEndpoints({
	endpoints: (build) => ({
		getPendingOffers: build.query<PendingOffer[], void>({
			queryFn: async (_arg, api, _options, baseQuery) => {
				const result = await api.dispatch(applicationApi.endpoints.getAllMyApplications.initiate(undefined, {
					subscribe: false, forceRefetch: true,
				}));
				if (result.isError || !result.data) return {
					error: { status: 'CUSTOM_ERROR', error: getErrorMessage(result.error, 'Не удалось загрузить полный список заявок.') },
				};
				const entries = result.data.filter(isPendingOffer);
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
