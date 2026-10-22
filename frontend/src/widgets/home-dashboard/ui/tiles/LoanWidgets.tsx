import { type DashboardWidget } from '@entities/dashboard';
import { loanOverview, paymentAmounts, useGetLoanQuery, useGetLoansSummaryQuery } from '@entities/loan';
import { getErrorStatus } from '@shared/api';
import { ROUTES } from '@shared/config';
import { formatDate, formatMoney, formatRate, formatTerm } from '@shared/lib/format';
import { Button } from '@shared/ui';
import { Link } from 'react-router';

import styles from '../Dashboard.module.scss';
import { WidgetState } from './WidgetState';

const freshness = { refetchOnFocus: true, refetchOnReconnect: true, refetchOnMountOrArgChange: true };
const creditHref = (id: string, closed: boolean) => `${ROUTES.credits}?loan=${encodeURIComponent(id)}&tab=${closed ? 'closed' : 'active'}`;

export const LoanWidget = ({ widget, onConfigure }: { widget: Extract<DashboardWidget, { type: 'loan' }>; onConfigure: () => void }) => {
	const { loanId } = widget.settings;
	const query = useGetLoanQuery(loanId ?? '', { ...freshness, skip: !loanId });
	if (!loanId) return <div className={styles.state}><p className={styles.muted}>Выберите договор для этого виджета.</p><Button variant="secondary" compact onClick={onConfigure}>Выбрать кредит</Button></div>;
	if (getErrorStatus(query.error) === 404) return <div className={styles.state}><p className={styles.muted}>Выбранный кредит больше недоступен.</p><Button variant="secondary" compact onClick={onConfigure}>Выбрать другой</Button></div>;
	const loan = query.currentData;
	const overview = loan ? loanOverview(loan) : null;
	const closed = overview?.closed ?? false;
	const percent = overview?.percent ?? null;
	const next = overview?.next;
	return <WidgetState loading={query.isLoading || (query.isFetching && !query.currentData)} error={query.isError} retry={() => void query.refetch()}>
		{loan && <>
			<p className={styles.muted}>{loan.product?.name ?? 'Кредит'} · {loan.contractNumber}{closed ? ' · закрыт' : ''}</p>
			{Boolean(overview?.overdue) && <p className={styles.muted}>Есть просроченные платежи. Проверьте график.</p>}
			<div><p className={styles.muted}>Остаток основного долга</p><div className={styles.number}>{loan.balance === undefined ? '—' : formatMoney(loan.balance)}</div></div>
			<div className={styles.inline}><span className={styles.muted}>{formatRate(loan.terms.rate)}</span><span className={styles.muted}>{formatTerm(loan.terms.term)}</span></div>
			{widget.h >= 4 && <p className={styles.muted}>{closed ? 'Кредит закрыт' : next ? `${formatDate(next.date)} · ${formatMoney(paymentAmounts(next).remaining)}` : 'Будущих платежей нет. Проверьте полный график.'}</p>}
			{percent !== null && <><div className={styles.progress} role="progressbar" aria-label="Погашено основного долга" aria-valuenow={percent} aria-valuemin={0} aria-valuemax={100}><span style={{ width: `${percent}%` }} /></div><p className={styles.muted}>Погашено {percent}% основного долга</p></>}
			<Link className={styles.link} to={creditHref(loanId, closed)}>Детали кредита ↗</Link>
		</>}
	</WidgetState>;
};

const AllPaymentsWidget = () => {
	const query = useGetLoansSummaryQuery(undefined, freshness);
	const payment = query.currentData?.nextPayment;
	return <WidgetState loading={query.isLoading || (query.isFetching && !query.currentData)} error={query.isError} retry={() => void query.refetch()} empty={!payment}
		emptyText="Будущих платежей нет. Просроченные обязательства смотрите в графике.">
		{payment && <><div className={styles.number}>{formatMoney(payment.amount)}</div><p className={styles.muted}>{formatDate(payment.date)} · по всем кредитам в эту дату</p><Link className={styles.link} to={ROUTES.creditsCalendar}>Платёжный календарь ↗</Link></>}
	</WidgetState>;
};
const SinglePaymentWidget = ({ id }: { id: string }) => {
	const query = useGetLoanQuery(id, freshness);
	const payment = query.currentData ? loanOverview(query.currentData).next : undefined;
	return <WidgetState loading={query.isLoading || (query.isFetching && !query.currentData)} error={query.isError} retry={() => void query.refetch()} empty={!payment}
		emptyText={query.currentData?.status === 'CLOSED' ? 'Выбранный кредит закрыт.' : 'Будущих платежей по выбранному кредиту нет.'}>
		{payment && <><div className={styles.number}>{formatMoney(paymentAmounts(payment).remaining)}</div><p className={styles.muted}>{formatDate(payment.date)} · {query.currentData?.contractNumber}</p><Link className={styles.link} to={creditHref(id, false)}>График кредита ↗</Link></>}
	</WidgetState>;
};
export const PaymentWidget = ({ widget }: { widget: Extract<DashboardWidget, { type: 'payment' }> }) =>
	widget.settings.loanId ? <SinglePaymentWidget id={widget.settings.loanId} /> : <AllPaymentsWidget />;

export const DebtWidget = () => {
	const query = useGetLoansSummaryQuery(undefined, freshness);
	const summary = query.currentData;
	return <WidgetState loading={query.isLoading || (query.isFetching && !query.currentData)} error={query.isError} retry={() => void query.refetch()}>
		<div className={styles.number}>{summary?.totalPrincipalBalance === undefined ? '—' : formatMoney(summary.totalPrincipalBalance)}</div>
		<p className={styles.muted}>{summary?.totalPrincipalBalance === undefined ? 'Сервер не вернул общий остаток основного долга.' : `Действующих кредитов: ${summary.activeLoans}. Только основной долг.`}</p>
		<Link className={styles.link} to={ROUTES.credits}>Мои кредиты ↗</Link>
	</WidgetState>;
};
