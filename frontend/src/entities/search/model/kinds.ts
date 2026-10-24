import { type SearchKind } from './schema';

export const SEARCH_KINDS: { type: SearchKind; label: string }[] = [
	{ type: 'applications', label: 'Заявки' },
	{ type: 'loans', label: 'Кредиты' },
	{ type: 'plans', label: 'Планы' },
	{ type: 'products', label: 'Продукты' },
	{ type: 'faq', label: 'Вопросы и ответы' },
	{ type: 'tickets', label: 'Обращения' },
];

export const normalizeSearch = (value: string) => value.toLocaleLowerCase('ru-RU')
	.replaceAll('ё', 'е').trim().replace(/\s+/gu, ' ');
