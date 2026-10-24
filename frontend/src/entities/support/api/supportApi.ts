import { baseApi } from '@shared/api';

import { type Faq, type FaqStep, type Ticket } from '../model/types';

export const supportApi = baseApi.injectEndpoints({
	endpoints: (build) => ({
		getFaqItem: build.query<Faq, string>({
			query: (id) => `/support/faq/${encodeURIComponent(id)}`,
		}),
		getTicket: build.query<Ticket, string>({
			query: (id) => `/support/tickets/${encodeURIComponent(id)}`,
			providesTags: [{ type: 'Ticket', id: 'LIST' }],
		}),
		getFaq: build.query<Faq[], FaqStep>({
			query: (step) => `/support/faq?step=${step}`,
		}),
		getTickets: build.query<Ticket[], void>({
			query: () => '/support/tickets',
			providesTags: [{ type: 'Ticket', id: 'LIST' }],
		}),
	}),
});

export const { useGetFaqItemQuery, useGetTicketQuery, useGetFaqQuery, useGetTicketsQuery } = supportApi;
