export type NotificationType = 'INFO_REQUESTED' | 'DECISION' | 'PAYMENT_DUE' | 'TICKET_REPLY';

export interface Notification {
	id: string;
	type: NotificationType;
	title: string;
	text: string;
	link: string;
	read: boolean;
	createdAt: string;
}

export interface NotificationsResponse {
	unreadCount: number;
	items: Notification[];
}
