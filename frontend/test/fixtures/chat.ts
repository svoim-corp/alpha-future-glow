import { type ChatMessage, type Conversation } from '@entities/support';

export const conversationFixture: Conversation = {
	id: 'chat-1', applicationId: null, createdAt: '2026-10-21T12:00:00Z',
};

export const assistantMessageFixture: ChatMessage = {
	id: 'm-1', role: 'assistant', status: 'COMPLETED', text: 'Вот ближайший платёж.',
	createdAt: '2026-10-21T12:00:00Z', sources: [{ id: 'kb-1', title: 'График платежей', href: '/support?tab=faq' }],
	blocks: [{ kind: 'payment', title: 'Ближайший платёж', payment: {
		loanId: 'l-1', productName: 'Оборотный кредит', date: '2026-11-01', amount: 100000,
	} }],
};
