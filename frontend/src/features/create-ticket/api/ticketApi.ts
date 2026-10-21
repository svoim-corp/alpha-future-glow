import { baseApi } from '@shared/api';

export interface CreateTicketArgs {
	subject: string;
	text: string;
	applicationId?: string;
}

export const ticketApi = baseApi.injectEndpoints({
	endpoints: (build) => ({
		createTicket: build.mutation<{ id: string; status: string; createdAt: string }, CreateTicketArgs>({
			query: (body) => ({ url: '/support/tickets', method: 'POST', body }),
			invalidatesTags: [{ type: 'Ticket', id: 'LIST' }],
		}),
	}),
});

export const { useCreateTicketMutation } = ticketApi;
