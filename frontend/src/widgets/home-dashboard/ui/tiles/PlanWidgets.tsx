import { type DashboardWidget } from '@entities/dashboard';
import { useGetScheduleQuery } from '@entities/loan';
import { formatPeriod, moscowToday, planCoversDate, useGetPlansQuery } from '@entities/plan';
import { ROUTES } from '@shared/config';
import { buildMonthGrid, monthRange, type MonthRef, shiftMonth } from '@shared/lib/calendar';
import { formatDate, formatMoney } from '@shared/lib/format';
import { Button, Modal, MonthPicker } from '@shared/ui';
import { useState } from 'react';
import { Link } from 'react-router';

import { calendarPayments } from '../../model/calendarPayments';
import styles from '../Dashboard.module.scss';
import { WidgetState } from './WidgetState';

const freshness = { refetchOnFocus: true, refetchOnReconnect: true, refetchOnMountOrArgChange: true };
export const PlansWidget = ({ widget }: { widget: Extract<DashboardWidget, { type: 'plans' }> }) => {
	const query = useGetPlansQuery('ACTIVE', freshness);
	const today = moscowToday();
	const plans = (query.data?.items ?? []).filter((plan) => widget.settings.filter === 'ideas'
		? plan.dateMode === 'UNSCHEDULED' : plan.dateMode !== 'UNSCHEDULED' && (plan.endDate ?? plan.startDate ?? '') >= today)
		.sort((a, b) => (a.startDate ?? '').localeCompare(b.startDate ?? '') || b.createdAt.localeCompare(a.createdAt));
	const limited = (query.data?.items.length ?? 0) >= 200;
	return <WidgetState loading={query.isLoading} error={query.isError} retry={() => void query.refetch()} empty={plans.length === 0}
		emptyText={limited ? 'В полученной выборке подходящих планов нет. Список ограничен 200 записями.' : widget.settings.filter === 'ideas' ? 'Идей без даты пока нет.' : 'Ближайших планов пока нет.'}>
		{limited && <p className={styles.notice}>Получено до 200 планов. Список может быть неполным.</p>}
		<div className={styles.planList}>{plans.slice(0, widget.h >= 5 ? 4 : 3).map((plan) => <Link to={ROUTES.planner} key={plan.id} className={styles.plan}>
			<div><strong>{plan.title}</strong><span className={styles.muted}>{formatPeriod(plan)}</span></div><span>{formatMoney(plan.budgetRub)}</span>
		</Link>)}</div>
		<Link className={styles.link} to={ROUTES.planner}>Все планы ↗</Link>
	</WidgetState>;
};

export const CalendarWidget = ({ widget }: { widget: Extract<DashboardWidget, { type: 'calendar' }> }) => {
	const [month, setMonth] = useState<MonthRef>(() => { const now = new Date(); return { year: now.getUTCFullYear(), month: now.getUTCMonth() }; });
	const [selected, setSelected] = useState<string | null>(null);
	const [pickingMonth, setPickingMonth] = useState(false);
	const showPayments = widget.settings.source !== 'plans';
	const showPlans = widget.settings.source !== 'payments';
	const range = monthRange(month);
	const payments = useGetScheduleQuery(range, { ...freshness, skip: !showPayments });
	const plans = useGetPlansQuery('ACTIVE', { ...freshness, skip: !showPlans });
	const paymentItems = calendarPayments(showPayments ? payments.currentData ?? [] : []);
	const planItems = showPlans ? plans.data?.items ?? [] : [];
	const loading = (showPayments && (payments.isLoading || (payments.isFetching && !payments.currentData))) || (showPlans && plans.isLoading);
	const error = (showPayments && payments.isError) || (showPlans && plans.isError);
	const monthLabel = new Intl.DateTimeFormat('ru-RU', { month: 'long', year: 'numeric', timeZone: 'UTC' }).format(new Date(Date.UTC(month.year, month.month, 1)));
	const today = new Date().toISOString().slice(0, 10);
	return <>
		<div className={styles.monthControls}>
			<Button compact variant="ghost" aria-label="Предыдущий месяц" onClick={() => setMonth((value) => shiftMonth(value, -1))}>‹</Button>
			<button className={styles.monthTitle} type="button" onClick={() => setPickingMonth(true)} aria-haspopup="dialog">{monthLabel}</button>
			<Button compact variant="ghost" aria-label="Следующий месяц" onClick={() => setMonth((value) => shiftMonth(value, 1))}>›</Button>
		</div>
		<WidgetState loading={loading} error={error} retry={() => { if (showPayments) void payments.refetch(); if (showPlans) void plans.refetch(); }}>
			{showPlans && planItems.length >= 200 && <p className={styles.notice}>Планы показаны по выборке до 200 записей.</p>}
			<div className={styles.calendar} role="group" aria-label={`События за ${monthLabel}`}>
				{['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'].map((day) => <span className={styles.weekday} key={day}>{day}</span>)}
				{buildMonthGrid(month).flatMap((week, weekIndex) => week.map((date, index) => {
					if (!date) return <span key={`empty-${weekIndex}-${index}`} />;
					const paymentCount = paymentItems.reduce((count, item) => count + (item.date === date ? item.count : 0), 0);
					const planCount = planItems.filter((plan) => planCoversDate(plan, date)).length;
					return <button key={date} type="button" aria-current={date === today ? 'date' : undefined}
						aria-label={`${formatDate(date)}: платежей ${paymentCount}, планов ${planCount}`} aria-haspopup="dialog" onClick={() => setSelected(date)}>
						{Number(date.slice(8))}<span className={styles.dots} aria-hidden="true">{paymentCount > 0 && <i />}{planCount > 0 && <i className={styles.planDot} />}</span>
					</button>;
				}))}
			</div>
			<div className={styles.legend}><span>● Платежи</span><span>● Планы</span></div>
		</WidgetState>
		{pickingMonth && <Modal title="Месяц событий" size="small" onClose={() => setPickingMonth(false)} footer={<Button onClick={() => setPickingMonth(false)}>Готово</Button>}>
			<MonthPicker value={month} onChange={setMonth} />
		</Modal>}
		{selected && <Modal title={`События ${formatDate(selected)}`} onClose={() => setSelected(null)}>
			<div className={styles.events}>
				{paymentItems.filter((item) => item.date === selected).map((item) => <Link key={item.key} to={`${ROUTES.credits}?loan=${encodeURIComponent(item.loanId)}`}>
					<strong>{item.productName}</strong><span>{formatMoney(item.amount)} · {item.count > 1 ? `платежей: ${item.count}` : 'платёж'} · {item.status === 'PAID' ? 'оплачено' : item.status === 'OVERDUE' ? 'просрочено' : 'запланировано'}</span>
				</Link>)}
				{planItems.filter((plan) => planCoversDate(plan, selected)).map((plan) => <Link key={plan.id} to={ROUTES.planner}><strong>{plan.title}</strong><span>{formatMoney(plan.budgetRub)} · план</span></Link>)}
				{!paymentItems.some((item) => item.date === selected) && !planItems.some((plan) => planCoversDate(plan, selected)) && <p>На эту дату событий нет.</p>}
			</div>
		</Modal>}
	</>;
};
