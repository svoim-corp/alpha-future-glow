export { supportApi, useGetFaqQuery, useGetTicketsQuery } from './api/supportApi';
export { useGetChatMessagesQuery,useGetConversationsQuery } from './api/supportApi';
export type { ChatAction,ChatBlock, ChatMessage, Conversation } from './model/chat';
export type { Faq, FaqStep, Ticket, TicketStatus } from './model/types';
export { FAQ_STEPS } from './model/types';
export { FaqList } from './ui/FaqList';
export { TicketList } from './ui/TicketList';
