import { type DashboardWidget } from '@entities/dashboard';
import { paymentGroups, upcomingPaymentRange, useGetLoanQuery, useGetLoansQuery, useGetScheduleQuery } from '@entities/loan';
import { getErrorStatus } from '@shared/api';
import { ROUTES } from '@shared/config';
import { formatDate, formatMoney, formatRate } from '@shared/lib/format';
import { Button } from '@shared/ui';
import { Link } from 'react-router';

import styles from '../Dashboard.module.scss';
import { WidgetState } from './WidgetState';

const freshness = { refetchOnFocus: true, refetchOnReconnect: true, refetchOnMountOrArgChange: true };
export const LoansWidget = ({ widget }: { widget: Extract<DashboardWidget, { type: 'loans' }> }) => {
	const query = useGetLoansQuery(undefined, freshness);
	const items = (query.currentData ?? []).filter((item) => item.status === widget.settings.status);
	return <WidgetState loading={query.isLoading} error={query.isError} retry={() => void query.refetch()}>
		<p className={styles.muted}>{widget.settings.status === 'ACTIVE' ? 'Действующие договоры' : 'Закрытые договоры'}</p>
		{items.length === 0 ? <p className={styles.empty}>Таких кредитов пока нет.</p>
			: <div className={styles.portfolio}>{items.slice(0, widget.h >= 4 ? 4 : 3).map((loan) => <Link key={loan.id} className={styles.portfolioCard} to={`${ROUTES.credits}?loan=${encodeURIComponent(loan.id)}&tab=${loan.status === 'CLOSED' ? 'closed' : 'active'}`}>
				<strong>{loan.product.name}</strong><span className={styles.muted}>{loan.id}</span>
				<span className={styles.smallNumber}>{formatMoney(loan.balance)}</span><span className={styles.muted}>Основной долг · {formatRate(loan.rate)}</span>
			</Link>)}</div>}
		<Link className={styles.link} to={`${ROUTES.credits}?tab=${widget.settings.status === 'CLOSED' ? 'closed' : 'active'}`}>Открыть кредиты ↗</Link>
	</WidgetState>;
};

export const PaymentsWidget = ({ widget, onConfigure }: { widget: Extract<DashboardWidget, { type: 'payments' }>; onConfigure: () => void }) => {
	const range = upcomingPaymentRange(widget.settings.days);
	const schedule = useGetScheduleQuery(range, freshness);
	const loan = useGetLoanQuery(widget.settings.loanId ?? '', { ...freshness, skip: widget.settings.loanId === null });
	const groups = paymentGroups(schedule.currentData ?? [], widget.settings.loanId);
	const total = groups.reduce((sum, group) => sum + group.amount, 0);
	const ceiling = Math.max(1, ...groups.map((group) => group.amount));
	const loanLoading = widget.settings.loanId !== null && (loan.isLoading || (loan.isFetching && !loan.currentData));
	if (widget.settings.loanId && getErrorStatus(loan.error) === 404) return <div className={styles.state}><p>Выбранный кредит недоступен.</p><Button variant="secondary" compact onClick={onConfigure}>Выбрать другой</Button></div>;
	return <WidgetState loading={schedule.isLoading || (schedule.isFetching && !schedule.currentData) || loanLoading}
		error={schedule.isError || (widget.settings.loanId !== null && loan.isError)} retry={() => { void schedule.refetch(); if (widget.settings.loanId) void loan.refetch(); }}>
		<div className={styles.inline}><span className={styles.smallNumber}>{formatMoney(total)}</span><span className={styles.muted}>{formatDate(range.from)} — {formatDate(range.to)}</span></div>
		<p className={styles.muted}>{widget.settings.loanId ? `По договору ${loan.currentData?.contractNumber ?? widget.settings.loanId}` : 'По всем кредитам'} · запланированные платежи</p>
		{groups.length === 0 ? <p className={styles.empty}>На этот период платежей не запланировано.</p> : <div className={styles.paymentBars} data-compact={widget.w < 8}>
			{groups.slice(0, widget.h >= 4 ? 7 : 3).map((group) => <Link className={styles.paymentBar} key={group.date} to={ROUTES.creditsCalendar}>
				<span>{formatDate(group.date)}</span><span className={styles.barTrack} aria-hidden="true"><i style={{ width: `${group.amount / ceiling * 100}%` }} /></span>
				<strong>{formatMoney(group.amount)}</strong><small>Платежей: {group.count}</small>
			</Link>)}
		</div>}
		<Link className={styles.link} to={ROUTES.creditsCalendar}>Полный календарь платежей ↗</Link>
	</WidgetState>;
};
