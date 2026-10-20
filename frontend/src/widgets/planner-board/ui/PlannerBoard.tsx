import { type BusinessPlan, formatPeriod, PlanCard, planCoversDate, useGetPlansQuery } from '@entities/plan';
import { useGetGoalsQuery } from '@entities/product';
import { DeletePlan } from '@features/manage-plans';
import { buildMonthGrid, type MonthRef, shiftMonth } from '@shared/lib/calendar';
import { formatMoney } from '@shared/lib/format';
import { AsyncBoundary, Button, Card } from '@shared/ui';
import clsx from 'clsx';
import { useState } from 'react';

import styles from './PlannerBoard.module.scss';

type Tab = 'calendar' | 'plans' | 'ideas';

const WEEKDAYS = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'];
const monthFormatter = new Intl.DateTimeFormat('ru-RU', { month: 'long', year: 'numeric' });

const startMonthOf = (plan: BusinessPlan | undefined): MonthRef => {
	const source = plan?.startDate ? new Date(`${plan.startDate}T00:00:00`) : new Date();

	return { year: source.getFullYear(), month: source.getMonth() };
};

export const PlannerBoard = () => {
	const { data, isLoading, isError, refetch } = useGetPlansQuery();
	const goals = useGetGoalsQuery();
	const [tab, setTab] = useState<Tab>('calendar');
	const [selectedId, setSelectedId] = useState<string | null>(null);
	const [month, setMonth] = useState<MonthRef | null>(null);

	const plans = data?.items ?? [];
	const dated = plans.filter((plan) => plan.dateMode !== 'UNSCHEDULED');
	const ideas = plans.filter((plan) => plan.dateMode === 'UNSCHEDULED');
	const selected = plans.find((plan) => plan.id === selectedId) ?? dated[0] ?? plans[0];
	const current = month ?? startMonthOf(selected);
	const goalName = (plan: BusinessPlan) => goals.data?.find((goal) => goal.code === plan.goal)?.name;

	const choose = (plan: BusinessPlan) => {
		setSelectedId(plan.id);
		setMonth(plan.startDate ? startMonthOf(plan) : null);
		setTab('calendar');
	};

	const cards = (list: BusinessPlan[]) => (
		<div className={styles.cards}>
			{list.map((plan) => (
				<PlanCard
					key={plan.id}
					plan={plan}
					goalName={goalName(plan)}
					actions={(
						<>
							<Button variant="secondary" compact onClick={() => choose(plan)}>{`Открыть: ${plan.title}`}</Button>
							<DeletePlan id={plan.id} title={plan.title} />
						</>
					)}
				/>
			))}
		</div>
	);

	return (
		<AsyncBoundary
			isLoading={isLoading}
			isError={isError}
			isEmpty={plans.length === 0}
			onRetry={() => void refetch()}
			loadingLabel="Загружаем планы"
			errorText="Не удалось загрузить планы"
			emptyTitle="Планов пока нет"
			emptyText="Запишите намерение — цель, бюджет и, если уже известно, даты. Когда план станет актуален, подходящие варианты появятся на главной."
		>
			<div className={styles.board}>
				{data && (
					<dl className={styles.metrics}>
						<div className={styles.metric}><dt>Планы развития</dt><dd>{data.summary.total}</dd></div>
						<div className={styles.metric}>
							<dt>Общий ориентир бюджета</dt>
							<dd>{formatMoney(data.summary.budgetRub)}</dd>
						</div>
						<div className={styles.metric}><dt>Идеи без даты</dt><dd>{data.summary.unscheduled}</dd></div>
					</dl>
				)}
				<div className={styles.tabs} role="tablist" aria-label="Представление планов">
					{([['calendar', 'Календарь'], ['plans', 'Все планы'], ['ideas', 'Идеи без даты']] as const).map(
						([value, label]) => (
							<button
								key={value}
								type="button"
								role="tab"
								id={`tab-${value}`}
								aria-selected={tab === value}
								aria-controls={`panel-${value}`}
								className={styles.tab}
								onClick={() => setTab(value)}
							>
								{label}
							</button>
						),
					)}
				</div>
				{tab === 'calendar' && (
					<div role="tabpanel" id="panel-calendar" aria-labelledby="tab-calendar" className={styles.board}>
						<div className={styles.workspace}>
							<Card aria-labelledby="pl-month">
								<div className={styles.nav}>
									<Button variant="secondary" compact onClick={() => setMonth(shiftMonth(current, -1))}>
										Предыдущий месяц
									</Button>
									<h2 id="pl-month" className={styles.month}>
										{monthFormatter.format(new Date(current.year, current.month, 1))}
									</h2>
									<Button variant="secondary" compact onClick={() => setMonth(shiftMonth(current, 1))}>
										Следующий месяц
									</Button>
								</div>
								<table className={styles.table}>
									<caption className="visually-hidden">Периоды планов по дням</caption>
									<thead>
										<tr>{WEEKDAYS.map((day) => <th key={day} scope="col">{day}</th>)}</tr>
									</thead>
									<tbody>
										{buildMonthGrid(current).map((week) => (
											<tr key={week.find(Boolean) ?? 'empty'}>
												{week.map((date, index) => {
													const plan = date
														? dated.find((item) => planCoversDate(item, date))
														: undefined;

													return (
														<td
															key={date ?? `e-${index}`}
															className={clsx(styles.day, plan && styles.covered)}
														>
															{date && (
																<span className={styles.dayLabel}>
																	{Number(date.slice(8))}
																	{plan && <span className="visually-hidden">{` — ${plan.title}`}</span>}
																</span>
															)}
														</td>
													);
												})}
											</tr>
										))}
									</tbody>
								</table>
							</Card>
							{selected && (
								<Card aria-label="Выбранный план">
									<h3>{selected.title}</h3>
									<p>{formatPeriod(selected)}</p>
									<p>{formatMoney(selected.budgetRub)}</p>
									{selected.note && <p>{selected.note}</p>}
									<p>
										{selected.dateMode === 'UNSCHEDULED'
											? 'Подбор можно показывать сразу.'
											: `Подбор начнётся с ${formatPeriod(selected).split(' — ')[0]}.`}
									</p>
								</Card>
							)}
						</div>
						<div className={styles.shortcuts}>
							{plans.map((plan) => (
								<Button
									key={plan.id}
									variant={plan.id === selected?.id ? 'primary' : 'secondary'}
									compact
									onClick={() => choose(plan)}
								>
									{plan.title}
								</Button>
							))}
						</div>
					</div>
				)}
				{tab === 'plans' && (
					<div role="tabpanel" id="panel-plans" aria-labelledby="tab-plans">{cards(plans)}</div>
				)}
				{tab === 'ideas' && (
					<div role="tabpanel" id="panel-ideas" aria-labelledby="tab-ideas">
						{ideas.length === 0 ? <p>Идей без даты пока нет.</p> : cards(ideas)}
					</div>
				)}
			</div>
		</AsyncBoundary>
	);
};
