import { baseApi } from '@shared/api';

import { type ChatMessage, type Conversation } from '../model/chat';
import { type Faq, type FaqStep, type Ticket } from '../model/types';

export const supportApi = baseApi.injectEndpoints({
	endpoints: (build) => ({
		getConversations: build.query<Conversation[], void>({
			query: () => '/support/conversations',
			providesTags: [{ type: 'SupportChat', id: 'LIST' }],
		}),
		getChatMessages: build.query<ChatMessage[], string>({
			query: (id) => `/support/conversations/${encodeURIComponent(id)}/messages`,
			providesTags: (_result, _error, id) => [{ type: 'SupportChat', id }],
		}),
		getFaqItem: build.query<Faq, string>({
			query: (id) => `/support/faq/${encodeURIComponent(id)}`,
		}),
		getTicket: build.query<Ticket, string>({
			query: (id) => `/support/tickets/${encodeURIComponent(id)}`,
			providesTags: (_result, _error, id) => [{ type: 'Ticket', id }],
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

export const {
	useGetFaqItemQuery, useGetTicketQuery, useGetFaqQuery, useGetTicketsQuery,
	useGetConversationsQuery, useGetChatMessagesQuery,
} = supportApi;
