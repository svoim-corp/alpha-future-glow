const dateFormatter = new Intl.DateTimeFormat('ru-RU', { day: '2-digit', month: '2-digit', year: 'numeric' });
const dateTimeFormatter = new Intl.DateTimeFormat('ru-RU', {
	day: '2-digit',
	month: '2-digit',
	year: 'numeric',
	hour: '2-digit',
	minute: '2-digit',
});

/** ISO 8601 → «20.10.2026» */
export const formatDate = (iso: string): string => dateFormatter.format(new Date(iso));

/** ISO 8601 → «20.10.2026, 15:57» (в часовом поясе браузера) */
export const formatDateTime = (iso: string): string => dateTimeFormatter.format(new Date(iso));
