import { type ApplicationShort, type ApplicationStatus } from '@entities/application';

/** Поясняющая строка под названием продукта (макет applications.html) */
export const STATUS_HINT: Record<ApplicationStatus, string> = {
	DRAFT: 'Черновик ещё не отправлен в банк',
	SUBMITTED: 'Заявка принята, скоро начнём проверку',
	DATA_CHECK: 'Проверяем данные компании',
	INFO_REQUESTED: 'Банк запросил дополнительную информацию',
	SCORING: 'Проверяем данные и оцениваем условия',
	ON_DECISION: 'Заявка передана на решение',
	APPROVED: 'Условия готовы. Можно перейти к оформлению',
	COUNTER_OFFER: 'Банк предложил другие условия — посмотрите их',
	REJECTED: 'Решение доступно в карточке заявки',
	SIGNING: 'Договор готов к подписанию',
	ISSUED: 'Кредит выдан',
	CLOSED: 'Заявка закрыта',
};

export const actionLabel = (item: ApplicationShort) => {
	if (item.status === 'INFO_REQUESTED') {
		return 'Дополнить';
	}

	return item.status === 'SIGNING' ? 'Подписать' : 'Посмотреть';
};

export type SortKey = 'date' | 'amount-desc' | 'amount-asc';
export type PeriodKey = 'all' | '7' | '30' | '90';

export interface ListFilters {
	query: string;
	status: ApplicationStatus | '';
	product: string;
	period: PeriodKey;
	sort: SortKey;
}

export const DEFAULT_FILTERS: ListFilters = { query: '', status: '', product: '', period: 'all', sort: 'date' };

const DAY = 86400000;

/** Применяет поиск, фильтры и сортировку; now передаётся снаружи, чтобы функция оставалась чистой */
export const applyFilters = (items: ApplicationShort[], filters: ListFilters, now: number) => {
	const query = filters.query.trim().toLowerCase();
	const result = items.filter((item) => {
		const text = `${item.product.name} ${item.number ?? ''}`.toLowerCase();
		const age = (now - new Date(item.updatedAt).getTime()) / DAY;

		return (!query || text.includes(query))
			&& (!filters.status || item.status === filters.status)
			&& (!filters.product || item.product.code === filters.product)
			&& (filters.period === 'all' || age <= Number(filters.period));
	});

	return result.sort((a, b) => {
		if (filters.sort === 'amount-desc') {
			return b.amount - a.amount;
		}

		if (filters.sort === 'amount-asc') {
			return a.amount - b.amount;
		}

		return b.updatedAt.localeCompare(a.updatedAt);
	});
};
