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
import { AsyncBoundary, Button, DataTable, Metrics, SelectField } from '@shared/ui';
import { useState } from 'react';
import { Link } from 'react-router';

import styles from './BankQueue.module.scss';

const PAGE_SIZE = 20;
const ANY = '';

const STATUS_OPTIONS: { value: string; label: string }[] = [
	{ value: ANY, label: 'Все статусы' },
	...(['SUBMITTED', 'DATA_CHECK', 'INFO_REQUESTED', 'SCORING', 'ON_DECISION'] as ApplicationStatus[])
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

const STAT_LABELS = [
	['SUBMITTED', 'Подано'],
	['DATA_CHECK', 'Проверка данных'],
	['INFO_REQUESTED', 'Запрошена информация'],
	['SCORING', 'Скоринг'],
	['ON_DECISION', 'На решении'],
	['decidedToday', 'Решено сегодня'],
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
					items={STAT_LABELS.map(([key, label]) => ({ label, value: data[key] }))}
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
			<div className={styles.filters}>
				<SelectField
					label="Статус"
					value={filters.status ?? ANY}
					options={STATUS_OPTIONS}
					onChange={(next) => change({ status: (next || undefined) as ApplicationStatus | undefined })}
				/>
				<SelectField
					label="Продукт"
					value={filters.product ?? ANY}
					options={PRODUCT_OPTIONS}
					onChange={(next) => change({ product: next || undefined })}
				/>
				<SelectField
					label="Риск"
					value={filters.risk ?? ANY}
					options={RISK_OPTIONS}
					onChange={(next) => change({ risk: (next || undefined) as RiskBand | undefined })}
				/>
				<SelectField
					label="Режим решения"
					value={filters.decisionMode ?? ANY}
					options={MODE_OPTIONS}
					onChange={(next) => change({ decisionMode: (next || undefined) as DecisionMode | undefined })}
				/>
			</div>
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
							rows={data.items}
							getRowKey={(row) => row.id}
							columns={[
								{
									key: 'number',
									header: 'Заявка',
									render: (row) => <Link to={paths.bankApplication(row.id)}>{row.number}</Link>,
								},
								{
									key: 'company',
									header: 'Компания',
									render: (row) => row.company.name,
								},
								{ key: 'product', header: 'Продукт', render: (row) => row.product.name },
								{
									key: 'amount',
									header: 'Сумма и срок',
									align: 'right',
									render: (row) => `${formatMoney(row.amount)}, ${formatTerm(row.term)}`,
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
										<>
											{row.score ?? '—'} <RiskBadge band={row.riskBand} />
										</>
									),
								},
								{ key: 'mode', header: 'Режим', render: (row) => <DecisionModeLabel mode={row.decisionMode} /> },
								{ key: 'assignee', header: 'Исполнитель', render: (row) => <AssignApplication item={row} /> },
								{ key: 'submitted', header: 'Подана', render: (row) => formatDateTime(row.submittedAt) },
							]}
						/>
						<nav className={styles.pager} aria-label="Страницы очереди">
							<Button
								variant="secondary"
								compact
								disabled={data.page <= 1}
								onClick={() => setFilters((prev) => ({ ...prev, page: prev.page - 1 }))}
							>
								Назад
							</Button>
							<span>{`Страница ${data.page} из ${pages}`}</span>
							<Button
								variant="secondary"
								compact
								disabled={data.page >= pages}
								onClick={() => setFilters((prev) => ({ ...prev, page: prev.page + 1 }))}
							>
								Вперёд
							</Button>
						</nav>
					</>
				)}
			</AsyncBoundary>
		</div>
	);
};
