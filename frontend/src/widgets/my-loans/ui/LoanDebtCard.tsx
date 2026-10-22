import { useGetLoanDebtQuery } from '@entities/loan';
import { formatDate, formatDateTime, formatMoney } from '@shared/lib/format';
import { Alert, AsyncBoundary, KeyValueList } from '@shared/ui';

import styles from './LoanPanel.module.scss';

export const LoanDebtCard = ({ loanId }: { loanId: string }) => {
	const debt = useGetLoanDebtQuery(loanId, { refetchOnFocus: true, refetchOnReconnect: true, refetchOnMountOrArgChange: true });
	const data = debt.data;
	return (
		<section className={styles.card} aria-labelledby={`debt-${loanId}`}>
			<div className={styles.heading}><div><h2 id={`debt-${loanId}`}>Задолженность на сегодня</h2><p>Основной долг и начисления по данным сервера</p></div></div>
			<AsyncBoundary isLoading={debt.isLoading} isError={debt.isError} onRetry={() => void debt.refetch()}
				loadingLabel="Загрузка задолженности" errorText="Не удалось загрузить задолженность.">
				{data && <>
					<div className={styles.largeMoney}>{formatMoney(data.totalOutstanding)}</div>
					<KeyValueList items={[
						{ label: 'Основной долг', value: formatMoney(data.principalOutstanding) },
						{ label: 'Начисленные проценты', value: formatMoney(data.interestOutstanding) },
						{ label: 'Пени', value: formatMoney(data.penaltiesOutstanding) },
						{ label: 'Комиссии', value: formatMoney(data.feesOutstanding) },
					]} />
					{data.overdue.total > 0 && <Alert tone="danger" title={`В том числе просрочено: ${formatMoney(data.overdue.total)}`}>
						{data.overdue.since && <p>С {formatDate(data.overdue.since)} · дней просрочки: {data.overdue.days}.</p>}
						Основной долг: {formatMoney(data.overdue.principal)}, проценты: {formatMoney(data.overdue.interest)}, пени: {formatMoney(data.overdue.penalties)}, комиссии: {formatMoney(data.overdue.fees)}.
					</Alert>}
					{data.unappliedFunds > 0 && <Alert tone="warning" title={`Ожидают распределения: ${formatMoney(data.unappliedFunds)}`}>
						Средства учтены отдельно и пока не уменьшают показанную задолженность.
					</Alert>}
					{data.prepaidInterest > 0 && <p className={styles.note}>Проценты, оплаченные заранее: {formatMoney(data.prepaidInterest)}. В задолженность не входят.</p>}
					<p className={styles.note}>Расчёт на {formatDateTime(data.asOf)}. Операционная дата: {formatDate(data.businessDate)} ({data.timeZone}). Будущие проценты графика в итог не включены. Просрочка уже входит в сумму задолженности.</p>
				</>}
			</AsyncBoundary>
		</section>
	);
};
