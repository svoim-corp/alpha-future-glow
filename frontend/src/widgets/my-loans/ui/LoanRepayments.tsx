import { type RepaymentStatus, useGetLoanRepaymentsQuery } from '@entities/loan';
import { baseApi } from '@shared/api';
import { formatDateTime, formatMoney } from '@shared/lib/format';
import { useAppDispatch } from '@shared/lib/store';
import { AsyncBoundary, Badge, Button, DataTable, KeyValueList } from '@shared/ui';
import { useState } from 'react';

import styles from './LoanPanel.module.scss';

const statuses = {
	PENDING: { label: 'Ожидает обработки', tone: 'warning' },
	PROCESSING: { label: 'Обрабатывается банком', tone: 'info' },
	SUCCEEDED: { label: 'Учтена', tone: 'success' },
	FAILED: { label: 'Отклонена', tone: 'danger' },
} as const satisfies Record<RepaymentStatus, { label: string; tone: string }>;

export const LoanRepayments = ({ loanId }: { loanId: string }) => {
	const [page, setPage] = useState(1);
	const dispatch = useAppDispatch();
	const query = useGetLoanRepaymentsQuery({ loanId, page }, { refetchOnFocus: true, refetchOnReconnect: true, refetchOnMountOrArgChange: true });
	const data = query.currentData;
	const pages = data ? Math.max(1, Math.ceil(data.total / data.size)) : 1;
	return (
		<section className={styles.card} aria-labelledby={`repayments-${loanId}`}>
			<div className={styles.heading}>
				<div><h2 id={`repayments-${loanId}`}>История операций</h2><p>Обычные и досрочные погашения, учтённые сервером</p></div>
				<Button variant="ghost" disabled={query.isFetching} onClick={() => dispatch(baseApi.util.invalidateTags(['Loan']))}>Обновить операции</Button>
			</div>
			<p className={styles.note}>Операции, совершённые отдельно в банке, пока не загружаются в кабинет. Это не полная банковская выписка.</p>
			<AsyncBoundary isLoading={query.isLoading || (query.isFetching && !data)} isError={query.isError}
				onRetry={() => void query.refetch()} loadingLabel="Загрузка операций" errorText="Не удалось загрузить историю операций."
				isEmpty={data?.total === 0} emptyTitle="Учтённых операций пока нет">
				{data && <>
					{data.items.length === 0 ? <p role="status" className={styles.note}>На этой странице операций нет. Вернитесь на предыдущую страницу.</p> : <DataTable
						caption="История операций по кредиту" hideCaption className={styles.table} rows={data.items} getRowKey={(operation) => operation.id} columns={[
							{ key: 'date', header: 'Запрошено', render: (operation) => formatDateTime(operation.requestedAt) },
							{ key: 'type', header: 'Операция', render: (operation) => operation.type === 'EARLY' ? 'Досрочное погашение' : 'Обычное погашение' },
							{ key: 'requested', header: 'Запрошено к списанию', align: 'right', render: (operation) => formatMoney(operation.requestedAmount) },
							{ key: 'posted', header: 'Учтено', align: 'right', render: (operation) => operation.status === 'SUCCEEDED' && operation.postedAmount !== null ? formatMoney(operation.postedAmount) : '—' },
							{ key: 'status', header: 'Статус', render: (operation) => <Badge tone={statuses[operation.status].tone}>{statuses[operation.status].label}</Badge> },
							{ key: 'details', header: 'Подробности', render: (operation) => <details className={styles.operationDetails}>
								<summary>Подробности операции</summary>
								<KeyValueList items={[
									{ label: 'Номер операции', value: operation.id },
									{ label: 'Счёт', value: operation.account.maskedNumber },
									{ label: 'Дата учёта', value: operation.postedAt ? formatDateTime(operation.postedAt) : 'Не учтена' },
									...(operation.status === 'SUCCEEDED' && operation.allocation ? [
										{ label: 'Основной долг', value: formatMoney(operation.allocation.principal) },
										{ label: 'Проценты', value: formatMoney(operation.allocation.interest) },
										{ label: 'Пени', value: formatMoney(operation.allocation.penalties) },
										{ label: 'Комиссии', value: formatMoney(operation.allocation.fees) },
										{ label: 'Не распределено', value: operation.unappliedAmount === null ? '—' : formatMoney(operation.unappliedAmount) },
									] : []),
									...(operation.failureReason ? [{ label: 'Причина отказа', value: operation.failureReason }] : []),
								]} />
							</details> },
						]} />}
					<nav className={styles.pagination} aria-label="Страницы истории операций">
						<Button variant="secondary" disabled={page <= 1 || query.isFetching} onClick={() => setPage((current) => Math.max(1, current - 1))}>Назад</Button>
						<span aria-live="polite">Страница {page} из {pages} · всего {data.total}</span>
						<Button variant="secondary" disabled={page >= pages || query.isFetching} onClick={() => setPage((current) => current + 1)}>Далее</Button>
					</nav>
				</>}
			</AsyncBoundary>
		</section>
	);
};
