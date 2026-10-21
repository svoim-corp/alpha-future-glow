import { type NotificationsResponse } from '@entities/notification';

export const notificationsFixture: NotificationsResponse = {
	unreadCount: 1,
	items: [
		{
			id: 'n-15',
			type: 'INFO_REQUESTED',
			title: 'Банк просит дополнить заявку',
			text: 'Уточните фактический адрес склада',
			link: '/applications/a-1024',
			read: false,
			createdAt: '2026-10-20T12:40:00Z',
		},
		{
			id: 'n-14',
			type: 'PAYMENT_DUE',
			title: 'Платёж через 3 дня',
			text: '62 200 ₽ 20 ноября',
			link: '/credits/calendar',
			read: true,
			createdAt: '2026-10-19T08:00:00Z',
		},
	],
};
