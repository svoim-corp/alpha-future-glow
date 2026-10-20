import { type Faq, type Ticket } from '@entities/support';

export const faqFixture: Faq[] = [
	{
		id: 'f-3',
		step: 'APPLICATION',
		question: 'Зачем выписка из 1С?',
		answer: 'Она подтверждает обороты и позволяет получить больший лимит',
	},
];

export const ticketsFixture: Ticket[] = [
	{
		id: 't-12',
		subject: 'Не загружается выписка',
		status: 'ANSWERED',
		reply: 'Выгрузите выписку в формате 1CClientBankExchange',
		applicationId: 'a-1031',
		updatedAt: '2026-10-20T11:30:00Z',
	},
];
