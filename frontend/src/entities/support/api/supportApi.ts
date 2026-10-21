import { baseApi } from '@shared/api';

import { type Faq, type FaqStep, type Ticket } from '../model/types';

export const supportApi = baseApi.injectEndpoints({
	endpoints: (build) => ({
		getFaq: build.query<Faq[], FaqStep>({
			query: (step) => `/support/faq?step=${step}`,
		}),
		getTickets: build.query<Ticket[], void>({
			query: () => '/support/tickets',
			providesTags: [{ type: 'Ticket', id: 'LIST' }],
		}),
	}),
});

export const { useGetFaqQuery, useGetTicketsQuery } = supportApi;
