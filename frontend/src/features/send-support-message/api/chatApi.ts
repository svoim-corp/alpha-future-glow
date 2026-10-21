import { type ChatMessage, type Conversation } from '@entities/support';
import { baseApi } from '@shared/api';

export const chatApi = baseApi.injectEndpoints({
	endpoints: (build) => ({
		createConversation: build.mutation<Conversation, { applicationId?: string }>({
			query: (body) => ({ url: '/support/conversations', method: 'POST', body }),
			invalidatesTags: [{ type: 'SupportChat', id: 'LIST' }],
		}),
		sendSupportMessage: build.mutation<ChatMessage, { conversationId: string; text: string; key: string }>({
			query: ({ conversationId, text, key }) => ({
				url: `/support/conversations/${encodeURIComponent(conversationId)}/messages`,
				method: 'POST', body: { text }, headers: { 'Idempotency-Key': key },
			}),
			invalidatesTags: (_result, _error, { conversationId }) => [{ type: 'SupportChat', id: conversationId }],
		}),
	}),
});

export const { useCreateConversationMutation, useSendSupportMessageMutation } = chatApi;
