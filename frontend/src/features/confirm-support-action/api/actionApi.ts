import { type ChatBlock } from '@entities/support';
import { baseApi } from '@shared/api';

const actionApi = baseApi.injectEndpoints({
	endpoints: (build) => ({
		confirmSupportAction: build.mutation<ChatBlock, { conversationId: string; actionId: string }>({
			query: ({ conversationId, actionId }) => ({
				url: `/support/conversations/${encodeURIComponent(conversationId)}/actions/${encodeURIComponent(actionId)}/confirm`,
				method: 'POST', headers: { 'Idempotency-Key': actionId },
			}),
			invalidatesTags: ['Application', 'Ticket', 'SupportChat', 'Plan'],
		}),
	}),
});

export const { useConfirmSupportActionMutation } = actionApi;
