import {
	type BusinessPlan, formatPeriodLong, moscowToday, PLAN_STATE_LABELS, PlanCard, plansOnDate,
	type PlanState, useGetPlansQuery,
} from '@entities/plan';
import { useGetGoalsQuery } from '@entities/product';
import { CreatePlan } from '@features/manage-plans';
import { type MonthRef, shiftMonth } from '@shared/lib/calendar';
import { formatMoney } from '@shared/lib/format';
import { AsyncBoundary, Badge, Icon, Metrics, panelId, SelectField, tabId, Tabs } from '@shared/ui';
import { useEffect, useState } from 'react';

import { calendarDateLabel, PlanCalendar } from './PlanCalendar';
import { PlanDetails } from './PlanDetails';
import styles from './PlannerBoard.module.scss';

type Tab = 'calendar' | 'plans' | 'ideas' | 'archive';
const states: PlanState[] = ['ACTIVE', 'PAUSED', 'DONE'];
const monthOf = (date: string): MonthRef => ({ year: Number(date.slice(0, 4)), month: Number(date.slice(5, 7)) - 1 });

export const PlannerBoard = () => {
	const [filter, setFilter] = useState<PlanState | ''>('');
	const [tab, setTab] = useState<Tab>('calendar');
	const queryState = tab === 'archive' ? 'ARCHIVED' : filter || undefined;
	const { currentData: data, isLoading, isFetching, isError, refetch } = useGetPlansQuery(queryState);
	const goals = useGetGoalsQuery();
	const [today, setToday] = useState(moscowToday);
	const [month, setMonth] = useState(() => monthOf(today));
	const [date, setDate] = useState<string | null>(today);
	const [selectedId, setSelectedId] = useState<string | null>(null);
	const plans = data?.items ?? [];
	const ideas = plans.filter((plan) => plan.dateMode === 'UNSCHEDULED');
	const selected = plans.find((plan) => plan.id === selectedId);
	const dayPlans = date ? plansOnDate(plans, date) : [];
	const goalName = (plan: BusinessPlan) => goals.data?.find((goal) => goal.code === plan.goal)?.name;

	useEffect(() => {
		const timer = setInterval(() => setToday(moscowToday()), 60_000);

		return () => clearInterval(timer);
	}, []);

	const chooseDate = (value: string) => {
		setDate(value);
		setMonth(monthOf(value));
		setSelectedId(null);
	};
	const details = selected ? (
		<PlanDetails key={selected.id} plan={selected} today={today}
			goalName={goalName(selected)} onClose={() => setSelectedId(null)} />
	) : null;
	const cards = (items: BusinessPlan[]) => items.length > 0 ? (
		<div className={styles.cards}>{items.map((plan) => (
			<PlanCard key={plan.id} plan={plan} goalName={goalName(plan)}
				onOpen={() => setSelectedId(plan.id)} actions={(
					<Badge tone="neutral">{PLAN_STATE_LABELS[plan.state]}</Badge>
				)} />
		))}</div>
	) : <p className={styles.empty}>{tab === 'archive' ? 'В архиве пока нет планов' : tab === 'ideas' ? 'Идей без даты пока нет' : 'В этом списке пока нет планов'}</p>;

	return (
		<div className={styles.board}>
			<Metrics label={tab === 'archive' ? 'Обзор архива' : 'Обзор планов'} items={[
				{ label: tab === 'archive' ? 'Планы в архиве' : 'Планы развития', value: data?.summary.total ?? '—', hint: data ? `${data.summary.withPeriod} с датами · ${data.summary.unscheduled} без даты` : '— с датами · — без даты' },
				{ label: 'Общий ориентир бюджета', value: data ? formatMoney(data.summary.budgetRub) : '—', hint: 'По выбранному состоянию, не кредитные обязательства', accent: true },
				{ label: 'Идеи без даты', value: data?.summary.unscheduled ?? '—', hint: 'Можно вернуться к ним позже' },
			]} />
			<div className={styles.toolbar}>
				<Tabs label="Представление планов" value={tab} onChange={(value) => { setTab(value); setSelectedId(null); }} items={[
					{ id: 'calendar', label: 'Календарь' },
					{ id: 'plans', label: 'Все планы', count: tab === 'archive' ? undefined : plans.length },
					{ id: 'ideas', label: 'Идеи без даты', count: tab === 'archive' ? undefined : ideas.length },
					{ id: 'archive', label: 'Архив', count: tab === 'archive' ? plans.length : undefined },
				]} />
				<div className={`${styles.stateFilter} ${tab === 'archive' ? styles.hiddenFilter : ''}`} aria-hidden={tab === 'archive'}>
					<SelectField disabled={tab === 'archive'}
						label="Состояние" value={filter}
						options={[{ value: '', label: 'Все, кроме архива' },
							...states.map((value) => ({ value, label: PLAN_STATE_LABELS[value] }))]}
						onChange={(value) => {
							setFilter(states.find((state) => state === value) ?? '');
							setSelectedId(null);
						}}
					/>
				</div>
			</div>
			<AsyncBoundary isLoading={isLoading || (isFetching && !data)} isError={isError} onRetry={() => void refetch()} loadingLabel="Загружаем планы" errorText="Не удалось загрузить планы">
				{plans.length === 0 && tab === 'calendar' && <p className={styles.empty}>{filter ? 'В выбранном состоянии планов пока нет' : 'Планов пока нет'}</p>}
				{tab === 'archive' && <p className={styles.empty}>Здесь хранятся архивированные планы. Можно вернуть план в работу, изменив его состояние. Безвозвратно удалённые планы сервер не сохраняет.</p>}
				<div role="tabpanel" id={panelId(tab)} aria-labelledby={tabId(tab)} className={styles.panel}>
					{tab === 'calendar' ? (
						<div className={styles.panel}>
							<PlanCalendar month={month} plans={plans} today={today}
								onMonthChange={(value) => { setMonth(value); setDate(null); setSelectedId(null); }}
								selectedDate={date} onSelect={chooseDate} onToday={() => chooseDate(today)}
								onMove={(delta) => {
									setMonth(shiftMonth(month, delta));
									setDate(null);
									setSelectedId(null);
								}} />
							<section className={styles.card} aria-label={date ? `Планы на ${calendarDateLabel(date)}` : 'Планы дня'}>
								<h2 className={styles.sectionTitle}>{date ? calendarDateLabel(date) : 'Выберите день'}</h2>
								{dayPlans.length > 0 ? <ul className={styles.dayPlans}>{dayPlans.map((plan) => <li key={plan.id}><button type="button" className={styles.shortcut} aria-label={`Открыть план: ${plan.title}`} aria-pressed={selectedId === plan.id} onClick={() => setSelectedId(plan.id)}><span><strong>{plan.title}</strong><small>{formatPeriodLong(plan)} · {formatMoney(plan.budgetRub)}</small></span><Icon name="arrow" size={16} /></button></li>)}</ul>
									: <p className={styles.empty}>{date ? 'На этот день планов нет' : 'Нажмите на дату в календаре, чтобы увидеть все её планы.'}</p>}
								{date && dayPlans.length === 0 && <CreatePlan date={date} label="Создать новый план" />}
							</section>
						</div>
					) : cards(tab === 'ideas' ? ideas : plans)}
				</div>
			</AsyncBoundary>
			{details}
		</div>
	);
};
