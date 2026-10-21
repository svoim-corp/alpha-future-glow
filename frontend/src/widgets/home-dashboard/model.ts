import { type ApplicationShort, type ApplicationStatus } from '@entities/application';

/** Три этапа рассмотрения на плитке «Мои заявки» (макет home.html) */
export const TRACK = ['Отправлена', 'Проверка данных', 'Решение'] as const;

/** Сколько этапов трека пройдено: 1 — отправлена, 2 — проверка данных, 3 — решение, 4 — всё пройдено */
export const trackStage = (status: ApplicationStatus): 1 | 2 | 3 | 4 => {
	if (status === 'SUBMITTED') {
		return 1;
	}

	if (status === 'DATA_CHECK' || status === 'INFO_REQUESTED') {
		return 2;
	}

	return status === 'SCORING' || status === 'ON_DECISION' ? 3 : 4;
};

export const isInProgress = (item: ApplicationShort) => item.status !== 'DRAFT' && item.status !== 'CLOSED';

/** Приветствие по времени суток */
export const greeting = (hour: number) => {
	if (hour < 5) {
		return 'Доброй ночи';
	}

	if (hour < 12) {
		return 'Доброе утро';
	}

	return hour < 18 ? 'Добрый день' : 'Добрый вечер';
};

/** Имя для приветствия: в ФИО «Фамилия Имя Отчество» имя — второе слово */
export const firstName = (fullName: string) => fullName.split(' ').filter(Boolean)[1] ?? fullName;

export const formatNumber = (value: number) => new Intl.NumberFormat('ru-RU').format(value);

const parse = (iso: string) => new Date(`${iso}T00:00:00`);

export const dateParts = (iso: string) => {
	const date = parse(iso);

	return {
		day: String(date.getDate()),
		month: new Intl.DateTimeFormat('ru-RU', { day: 'numeric', month: 'long' }).format(date).replace(/^\d+\s*/, ''),
		year: String(date.getFullYear()),
		weekday: new Intl.DateTimeFormat('ru-RU', { weekday: 'long' }).format(date),
		shortMonth: new Intl.DateTimeFormat('ru-RU', { month: 'short' }).format(date).replace('.', '').slice(0, 3)
			.toUpperCase(),
	};
};
