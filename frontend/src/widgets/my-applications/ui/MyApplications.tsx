import {
	applicationHref,
	type ApplicationShort,
	type ApplicationStatus,
	STATUS_VIEW,
	StatusBadge,
	useGetMyApplicationsQuery,
} from '@entities/application';
import { DeleteDraftButton } from '@features/delete-draft';
import { ROUTES } from '@shared/config';
import { formatDate, formatMoney, formatTerm } from '@shared/lib/format';
import {
	AsyncBoundary,
	Badge,
	Button,
	ButtonLink,
	FilterToolbar,
	Icon,
	ListToolbar,
	Metrics,
	panelId,
	SearchField,
	Select,
	SelectFilter,
	tabId,
	Tabs,
} from '@shared/ui';
import { useState } from 'react';
import { Link } from 'react-router';

import { actionLabel, applyFilters, DEFAULT_FILTERS, type ListFilters, STATUS_HINT } from '../model';
import styles from './MyApplications.module.scss';

type TabKey = 'submitted' | 'drafts';

const SUBMITTED_STATUSES = Object.keys(STATUS_VIEW).filter((status) => status !== 'DRAFT') as ApplicationStatus[];

const PERIOD_OPTIONS = [
	{ value: 'all', label: 'За всё время' },
	{ value: '7', label: 'За 7 дней' },
	{ value: '30', label: 'За 30 дней' },
	{ value: '90', label: 'За 90 дней' },
];

const SORT_OPTIONS = [
	{ value: 'date', label: 'Сначала новые' },
	{ value: 'amount-desc', label: 'По сумме: больше' },
	{ value: 'amount-asc', label: 'По сумме: меньше' },
];

const statusOptions = [
	{ value: '', label: 'Все статусы' },
	...SUBMITTED_STATUSES.map((status) => ({ value: status, label: STATUS_VIEW[status].label })),
];

const Row = ({ item }: { item: ApplicationShort }) => (
	<li className={styles.row}>
		<div className={styles.rowIcon}><Icon name="applications" /></div>
		<div className={styles.info}>
			<span className={styles.eyebrow}>{item.number ?? '—'} · {formatDate(item.updatedAt)}</span>
			<h3>{item.product.name}</h3>
			<p>{STATUS_HINT[item.status]}</p>
		</div>
		<div className={styles.finance}>
			<strong>{formatMoney(item.amount)}</strong>
			<small>{formatTerm(item.term)}</small>
		</div>
		<div className={styles.state}>
			<StatusBadge status={item.status} />
			<Link to={applicationHref(item)} className={styles.inlineAction}>
				{actionLabel(item)} <Icon name="arrow" size={14} />
			</Link>
		</div>
	</li>
);

const Draft = ({ item }: { item: ApplicationShort }) => (
	<li className={styles.draft}>
		<div className={styles.draftTop}>
			<Icon name="applications" />
			<Badge>Черновик</Badge>
		</div>
		<h3>{item.product.name}</h3>
		<strong className={styles.draftAmount}>{formatMoney(item.amount)}</strong>
		<p>{formatTerm(item.term)} · сохранён {formatDate(item.updatedAt)}</p>
		<div className={styles.draftActions}>
			<ButtonLink to={applicationHref(item)}>Продолжить заполнение <Icon name="arrow" size={17} /></ButtonLink>
			<DeleteDraftButton applicationId={item.id} productName={item.product.name} />
		</div>
	</li>
);

/** «Мои заявки» по макету applications.html: обзор, вкладки «Заявки» и «Черновики», поиск, фильтры и сортировка */
export const MyApplications = () => {
	const { data, isLoading, isError, refetch } = useGetMyApplicationsQuery();
	const [tab, setTab] = useState<TabKey>('submitted');
	const [filters, setFilters] = useState<ListFilters>(DEFAULT_FILTERS);
	const items = data?.items ?? [];
	const submitted = items.filter((item) => item.status !== 'DRAFT');
	const drafts = items.filter((item) => item.status === 'DRAFT');
	const shown = applyFilters(submitted, filters, new Date().getTime());
	const products = [...new Map(submitted.map((item) => [item.product.code, item.product.name])).entries()];
	const productOptions = [{ value: '', label: 'Все продукты' }, ...products.map(([value, label]) => ({ value, label }))];
	const set = <K extends keyof ListFilters>(key: K, value: ListFilters[K]) => {
		setFilters({ ...filters, [key]: value });
	};

	return (
		<AsyncBoundary
			isLoading={isLoading}
			isError={isError}
			isEmpty={items.length === 0}
			onRetry={() => void refetch()}
			loadingLabel="Загрузка заявок"
			errorText="Не удалось загрузить заявки."
			emptyTitle="Заявок пока нет"
			emptyText="Подберите продукт в каталоге — черновик сохранится, и вернуться к нему можно в любой момент."
			emptyAction={<ButtonLink to={ROUTES.products}>Получить финансирование</ButtonLink>}
		>
			<div className={styles.root}>
				<Metrics
					label="Обзор заявок"
					items={[
						{ label: 'Всего заявок', value: submitted.length, hint: 'Без учёта черновиков' },
						{
							label: 'Требуют решения',
							value: submitted.filter((item) => item.actionRequired).length,
							hint: 'Запрос банка и новые условия',
							accent: true,
						},
						{
							label: 'Одобрено',
							value: submitted.filter((item) => item.status === 'APPROVED').length,
							hint: 'Предложение готово к оформлению',
						},
					]}
				/>
				<Tabs
					label="Заявки и черновики"
					value={tab}
					onChange={setTab}
					items={[
						{ id: 'submitted', label: 'Заявки', count: submitted.length },
						{ id: 'drafts', label: 'Черновики', count: drafts.length },
					]}
				/>
				{tab === 'submitted' ? (
					<section id={panelId('submitted')} role="tabpanel" aria-labelledby={tabId('submitted')}>
						<div className={styles.card}>
							<FilterToolbar>
								<SearchField
									label="Поиск по заявкам"
									placeholder="Название или номер заявки"
									value={filters.query}
									onChange={(value) => set('query', value)}
								/>
								<SelectFilter
									label="Статус"
									value={filters.status}
									options={statusOptions}
									onChange={(next) => set('status', next as ApplicationStatus | '')}
								/>
								<SelectFilter
									label="Продукт"
									value={filters.product}
									options={productOptions}
									onChange={(next) => set('product', next)}
								/>
								<SelectFilter
									label="Период"
									value={filters.period}
									options={PERIOD_OPTIONS}
									onChange={(next) => set('period', next as ListFilters['period'])}
								/>
							</FilterToolbar>
							<ListToolbar>
								<span role="status">Найдено: {shown.length} из {submitted.length}</span>
								<span className={styles.sort}>
									Сортировать:
									<Select
										variant="inline"
										aria-label="Сортировка заявок"
										value={filters.sort}
										options={SORT_OPTIONS}
										onChange={(next) => set('sort', next as ListFilters['sort'])}
									/>
								</span>
							</ListToolbar>
							{shown.length === 0 ? (
								<div className={styles.empty}>
									<h3>Таких заявок пока нет</h3>
									<p>Попробуйте изменить запрос или сбросить фильтры.</p>
									<Button variant="secondary" onClick={() => setFilters(DEFAULT_FILTERS)}>Сбросить фильтры</Button>
								</div>
							) : (
								<ul className={styles.list} aria-label="Заявки">
									{shown.map((item) => <Row key={item.id} item={item} />)}
								</ul>
							)}
							<div className={styles.bottom}>
								<span>{shown.length === submitted.length ? 'Показаны все заявки' : 'Показаны найденные заявки'}</span>
								<span>{shown.length} из {submitted.length}</span>
							</div>
						</div>
					</section>
				) : (
					<section id={panelId('drafts')} role="tabpanel" aria-labelledby={tabId('drafts')}>
						<div className={styles.info2}>
							<Icon name="applications" />
							<p>Продолжите с того места, где остановились. Черновики ещё не отправлены в банк.</p>
						</div>
						{drafts.length === 0 ? (
							<p className={styles.emptyText}>Черновиков нет.</p>
						) : (
							<ul className={styles.drafts} aria-label="Черновики">
								{drafts.map((item) => <Draft key={item.id} item={item} />)}
							</ul>
						)}
					</section>
				)}
			</div>
		</AsyncBoundary>
	);
};
