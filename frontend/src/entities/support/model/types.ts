export type FaqStep = 'PRODUCT' | 'APPLICATION' | 'DOCUMENTS' | 'STATUS' | 'DECISION';

export interface Faq {
	id: string;
	step: FaqStep;
	question: string;
	answer: string;
}

export type TicketStatus = 'OPEN' | 'ANSWERED' | 'CLOSED';

export interface Ticket {
	id: string;
	subject: string;
	status: TicketStatus;
	reply: string | null;
	applicationId: string | null;
	updatedAt: string;
}

export const FAQ_STEPS: { value: FaqStep; label: string }[] = [
	{ value: 'PRODUCT', label: 'Выбор продукта' },
	{ value: 'APPLICATION', label: 'Заявка' },
	{ value: 'DOCUMENTS', label: 'Документы' },
	{ value: 'STATUS', label: 'Статус' },
	{ value: 'DECISION', label: 'Решение' },
];
