import {
	type ApplicationStatus,
	BANK_STATUS_LABEL,
	type BankQueueFilters,
	type BankQueueItem,
	type DecisionMode,
	DecisionModeLabel,
	RiskBadge,
	type RiskBand,
	StatusBadge,
	useGetBankApplicationsQuery,
	useGetBankStatsQuery,
} from '@entities/application';
import { AssignApplication } from '@features/assign-application';
import { paths } from '@shared/config';
import { formatDateTime, formatMoney, formatTerm } from '@shared/lib/format';
import { AsyncBoundary, Button, DataTable, FilterToolbar, ListToolbar, Metrics, SelectFilter } from '@shared/ui';
import clsx from 'clsx';
import { useState } from 'react';
import { Link } from 'react-router';

import styles from './BankQueue.module.scss';

const PAGE_SIZE = 20;
const ANY = '';

const STATUS_OPTIONS: { value: string; label: string }[] = [
	{ value: ANY, label: 'Все статусы' },
	...([
		'SUBMITTED', 'DATA_CHECK', 'INFO_REQUESTED', 'SCORING', 'ON_DECISION',
		'APPROVED', 'COUNTER_OFFER', 'REJECTED', 'SIGNING', 'ISSUED', 'CLOSED',
	] as ApplicationStatus[])
		.map((value) => ({ value, label: BANK_STATUS_LABEL[value] })),
];
const PRODUCT_OPTIONS = [
	{ value: ANY, label: 'Все продукты' },
	{ value: 'WORKING', label: 'Оборотный кредит' },
	{ value: 'INVEST', label: 'Инвестиционный кредит' },
	{ value: 'EXPRESS', label: 'Экспресс-кредит' },
	{ value: 'OVERDRAFT', label: 'Овердрафт' },
];
const RISK_OPTIONS = [
	{ value: ANY, label: 'Любой риск' },
	{ value: 'LOW', label: 'Низкий' },
	{ value: 'MEDIUM', label: 'Средний' },
	{ value: 'HIGH', label: 'Высокий' },
];
const MODE_OPTIONS = [
	{ value: ANY, label: 'Любой режим' },
	{ value: 'AUTO', label: 'Автоматически' },
	{ value: 'MANUAL', label: 'Ручной разбор' },
];

const STATS = [
	{ key: 'ON_DECISION', label: 'На решении', hint: 'Ждут решения сотрудника', accent: true },
	{ key: 'SUBMITTED', label: 'Подано', hint: 'Новые заявки' },
	{ key: 'DATA_CHECK', label: 'Проверка данных', hint: 'Идут внешние проверки' },
	{ key: 'INFO_REQUESTED', label: 'Запрошена информация', hint: 'Ждём ответ клиента' },
	{ key: 'SCORING', label: 'Скоринг', hint: 'Идёт оценка' },
	{ key: 'decidedToday', label: 'Решено сегодня', hint: 'С начала дня' },
] as const;

const Stats = () => {
	const { data, isLoading, isError, refetch } = useGetBankStatsQuery();

	return (
		<AsyncBoundary
			isLoading={isLoading}
			isError={isError}
			onRetry={() => void refetch()}
			loadingLabel="Загружаем счётчики"
			errorText="Не удалось загрузить счётчики"
		>
			{data && (
				<Metrics
					label="Заявки по этапам"
					items={STATS.map(({ key, ...item }) => ({ ...item, value: data[key] }))}
				/>
			)}
		</AsyncBoundary>
	);
};

export const BankQueue = () => {
	const [filters, setFilters] = useState<BankQueueFilters>({ page: 1, size: PAGE_SIZE });
	const { data, isLoading, isError, refetch } = useGetBankApplicationsQuery(filters);
	const pages = data ? Math.max(1, Math.ceil(data.total / data.size)) : 1;

	const change = (patch: Partial<BankQueueFilters>) => setFilters((prev) => ({ ...prev, ...patch, page: 1 }));

	return (
		<div className={styles.queue}>
			<Stats />
			<div className={styles.card}>
				<FilterToolbar>
					<SelectFilter
						label="Статус"
						value={filters.status ?? ANY}
						options={STATUS_OPTIONS}
						onChange={(next) => change({ status: (next || undefined) as ApplicationStatus | undefined })}
					/>
					<SelectFilter
						label="Продукт"
						value={filters.product ?? ANY}
						options={PRODUCT_OPTIONS}
						onChange={(next) => change({ product: next || undefined })}
					/>
					<SelectFilter
						label="Риск"
						value={filters.risk ?? ANY}
						options={RISK_OPTIONS}
						onChange={(next) => change({ risk: (next || undefined) as RiskBand | undefined })}
					/>
					<SelectFilter
						label="Режим решения"
						value={filters.decisionMode ?? ANY}
						options={MODE_OPTIONS}
						onChange={(next) => change({ decisionMode: (next || undefined) as DecisionMode | undefined })}
					/>
				</FilterToolbar>
				<ListToolbar>
					<span role="status">{data ? `Найдено: ${data.total}` : 'Загружаем заявки'}</span>
				</ListToolbar>
				<AsyncBoundary
					isLoading={isLoading}
					isError={isError}
					isEmpty={data?.items.length === 0}
					onRetry={() => void refetch()}
					loadingLabel="Загружаем заявки"
					errorText="Не удалось загрузить очередь"
					emptyTitle="По этим условиям заявок нет"
					emptyText="Измените фильтры, чтобы увидеть другие заявки."
				>
					{data && (
						<>
							<DataTable<BankQueueItem>
								caption="Очередь заявок"
								hideCaption
								className={styles.table}
								rows={data.items}
								getRowKey={(row) => row.id}
								columns={[
									{
										key: 'number',
										header: 'Заявка',
										render: (row) => (
											<span className={styles.cell}>
												<Link to={paths.bankApplication(row.id)} className={styles.link}>
													{row.number}
												</Link>
												<small>{formatDateTime(row.submittedAt)}</small>
											</span>
										),
									},
									{
										key: 'company',
										header: 'Компания',
										render: (row) => (
											<span className={styles.cell}>
												<strong>{row.company.name}</strong>
												<small>{row.product.name}</small>
											</span>
										),
									},
									{
										key: 'amount',
										header: 'Сумма и срок',
										align: 'right',
										render: (row) => (
											<span className={clsx(styles.cell, styles.end)}>
												<strong>{formatMoney(row.amount)}</strong>
												<small>{formatTerm(row.term)}</small>
											</span>
										),
									},
									{
										key: 'status',
										header: 'Статус',
										render: (row) => (
											<StatusBadge status={row.status} label={BANK_STATUS_LABEL[row.status]} />
										),
									},
									{
										key: 'score',
										header: 'Балл и риск',
										render: (row) => (
											<span className={styles.score}>
												<strong>{row.score ?? '—'}</strong>
												<RiskBadge band={row.riskBand} />
											</span>
										),
									},
									{ key: 'mode', header: 'Режим', render: (row) => <DecisionModeLabel mode={row.decisionMode} /> },
									{ key: 'assignee', header: 'Исполнитель', render: (row) => <AssignApplication item={row} /> },
								]}
							/>
							<nav className={styles.pager} aria-label="Страницы очереди">
								<span>{`Страница ${data.page} из ${pages}`}</span>
								<span className={styles.pagerButtons}>
									<Button
										variant="secondary"
										compact
										disabled={data.page <= 1}
										onClick={() => setFilters((prev) => ({ ...prev, page: prev.page - 1 }))}
									>
										Назад
									</Button>
									<Button
										variant="secondary"
										compact
										disabled={data.page >= pages}
										onClick={() => setFilters((prev) => ({ ...prev, page: prev.page + 1 }))}
									>
										Вперёд
									</Button>
								</span>
							</nav>
						</>
					)}
				</AsyncBoundary>
			</div>
		</div>
	);
};
