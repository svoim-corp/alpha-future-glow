import { type Conversation } from '@entities/support';

const dayKey = (date: Date) => (
	`${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
);

export const groupConversations = (items: Conversation[], now = new Date()) => {
	const today = dayKey(now);
	const yesterday = new Date(now);
	yesterday.setDate(yesterday.getDate() - 1);
	const previousDay = dayKey(yesterday);
	const groups = new Map<string, { key: string; label: string; items: Conversation[] }>();
	for (const item of items) {
		const date = new Date(item.createdAt);
		const key = dayKey(date);
		const label = key === today ? 'Сегодня' : key === previousDay ? 'Вчера' : date.toLocaleDateString('ru-RU');
		const group = groups.get(key) ?? { key, label, items: [] };
		group.items.push(item);
		groups.set(key, group);
	}
	return [...groups.values()].sort((a, b) => b.key.localeCompare(a.key));
};
