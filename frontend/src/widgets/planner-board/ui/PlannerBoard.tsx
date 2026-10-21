import {
	type BusinessPlan,
	formatPeriodLong,
	formatPlanStart,
	PlanCard,
	planCoversDate,
	planIcon,
	useGetPlansQuery,
} from '@entities/plan';
import { useGetGoalsQuery } from '@entities/product';
import { DeletePlan } from '@features/manage-plans';
import { ROUTES } from '@shared/config';
import { buildMonthCells, type MonthCell, type MonthRef, shiftMonth, toIsoDate } from '@shared/lib/calendar';
import { formatCompactMoney, formatMoney } from '@shared/lib/format';
import { AsyncBoundary, Badge, Button, Icon, Metrics, panelId, tabId, Tabs } from '@shared/ui';
import clsx from 'clsx';
import { useState } from 'react';
import { Link } from 'react-router';

import styles from './PlannerBoard.module.scss';

type Tab = 'calendar' | 'plans' | 'ideas';

const WEEKDAYS = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'];
const monthFormat = new Intl.DateTimeFormat('ru-RU', { month: 'long', year: 'numeric' });
/** «июнь 2027 г.» → «июнь 2027»: сокращение года в заголовке лишнее */
const monthFormatter = { format: (date: Date) => monthFormat.format(date).replace(/\s*г\.$/, '') };

const startMonthOf = (plan: BusinessPlan | undefined, today: Date): MonthRef => {
	const source = plan?.startDate ? new Date(`${plan.startDate}T00:00:00`) : today;

	return { year: source.getFullYear(), month: source.getMonth() };
};

const todayIso = (today: Date) => toIsoDate(today.getFullYear(), today.getMonth(), today.getDate());

const planState = (plan: BusinessPlan, today: string) => {
	if (plan.dateMode === 'UNSCHEDULED' || !plan.startDate) {
		return { label: 'Идея без даты', tone: 'neutral' as const };
	}

	return plan.startDate > today
		? { label: 'На будущее', tone: 'info' as const }
		: { label: 'Идёт сейчас', tone: 'success' as const };
};

const shortcutHint = (plan: BusinessPlan) => (
	`${plan.dateMode === 'UNSCHEDULED' || !plan.startDate ? 'Без даты' : formatPeriodLong(plan)} · ${formatCompactMoney(plan.budgetRub)}`
);

interface DayCellProps {
	cell: MonthCell;
	plan: BusinessPlan | undefined;
	isWeekStart: boolean;
	isToday: boolean;
}

/** День календаря: период плана подсвечен, подпись плана — в первый день и в начале каждой недели */
const DayCell = ({ cell, plan, isWeekStart, isToday }: DayCellProps) => (
	<td className={clsx(styles.day, plan && styles.covered, cell.outside && styles.outside)}>
		<span className={clsx(styles.dayLabel, isToday && styles.todayMark)}>
			{Number(cell.date.slice(8))}
			{plan && !cell.outside && <span className="visually-hidden">{` — ${plan.title}`}</span>}
		</span>
		{plan && (isWeekStart || cell.date === plan.startDate) && (
			<span className={styles.event} aria-hidden="true">{plan.title.split(' ')[0]}</span>
		)}
	</td>
);

const STEPS = [
	{ title: 'Запишите намерение', text: 'Задача, бюджет и даты — если они уже есть.' },
	{ title: 'Получайте подбор вовремя', text: 'Для плана с периодом — с его начала. Для идеи без даты — сразу.' },
	{
		title: 'Выбирайте следующий шаг',
		text: 'Сравните варианты на главной. Оформление начнётся только по вашему действию.',
	},
];

/** Планер по макету planner.html: обзор, вкладки, календарь с выбранным планом, «На горизонте», как это работает */
export const PlannerBoard = () => {
	const { data, isLoading, isError, refetch } = useGetPlansQuery();
	const goals = useGetGoalsQuery();
	const [tab, setTab] = useState<Tab>('calendar');
	const [selectedId, setSelectedId] = useState<string | null>(null);
	const [month, setMonth] = useState<MonthRef | null>(null);

	const now = new Date();
	const today = todayIso(now);
	const plans = data?.items ?? [];
	const dated = plans.filter((plan) => plan.dateMode !== 'UNSCHEDULED');
	const ideas = plans.filter((plan) => plan.dateMode === 'UNSCHEDULED');
	const selected = plans.find((plan) => plan.id === selectedId) ?? dated[0] ?? plans[0];
	const current = month ?? startMonthOf(selected, now);
	const goalName = (plan: BusinessPlan) => goals.data?.find((goal) => goal.code === plan.goal)?.name;
	const weeks = buildMonthCells(current);
	const planOn = (date: string) => dated.find((plan) => planCoversDate(plan, date));
	const coveredDays = weeks.flat().filter((cell) => !cell.outside && planOn(cell.date)).length;
	const selectedState = selected ? planState(selected, today) : null;

	const choose = (plan: BusinessPlan) => {
		setSelectedId(plan.id);
		setMonth(plan.startDate ? startMonthOf(plan, now) : null);
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
							<Button variant="ghost" compact onClick={() => choose(plan)} aria-label={`Открыть: ${plan.title}`}>
								Открыть в календаре <Icon name="arrow" size={14} />
							</Button>
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
				<section className={styles.intro} aria-label="О планере">
					<span className={styles.introIcon}><Icon name="calendar" size={22} /></span>
					<div>
						<h2>У каждого плана — своё время</h2>
						<p>
							Укажите период и бюджет или просто запишите идею. Когда план станет актуален,
							подходящие варианты появятся на главной.
						</p>
					</div>
					<Link to={ROUTES.home} className={styles.introLink}>Как это выглядит <Icon name="arrow" size={16} /></Link>
				</section>
				{data && (
					<Metrics
						label="Обзор планов"
						items={[
							{
								label: 'Планы развития',
								value: data.summary.total,
								hint: `${data.summary.withPeriod} с периодом · ${data.summary.unscheduled} без даты`,
							},
							{
								label: 'Общий ориентир бюджета',
								value: formatMoney(data.summary.budgetRub),
								hint: 'Пожелания, а не кредитные обязательства',
								accent: true,
							},
							{
								label: 'Идеи без даты',
								value: data.summary.unscheduled,
								hint: 'Подбор можно показывать уже сейчас',
							},
						]}
					/>
				)}
				<Tabs
					label="Представление планов"
					value={tab}
					onChange={setTab}
					items={[
						{ id: 'calendar', label: 'Календарь' },
						{ id: 'plans', label: 'Все планы', count: plans.length },
						{ id: 'ideas', label: 'Идеи без даты', count: ideas.length },
					]}
				/>
				{tab === 'calendar' && (
					<div role="tabpanel" id={panelId('calendar')} aria-labelledby={tabId('calendar')} className={styles.panel}>
						<div className={styles.workspace}>
							<section className={styles.card} aria-labelledby="pl-month">
								<div className={styles.calendarHead}>
									<div>
										<span className={styles.label}>ПЕРИОД ВЫБРАННОГО ПЛАНА</span>
										<h2 id="pl-month" className={styles.month}>
											{monthFormatter.format(new Date(current.year, current.month, 1))}
										</h2>
									</div>
									<div className={styles.controls}>
										<button
											type="button"
											className={styles.today}
											onClick={() => setMonth(startMonthOf(undefined, now))}
										>
											Сегодня
										</button>
										<button
											type="button"
											className={clsx(styles.arrow, styles.prev)}
											aria-label="Предыдущий месяц"
											onClick={() => setMonth(shiftMonth(current, -1))}
										>
											<Icon name="chevron" size={17} />
										</button>
										<button
											type="button"
											className={clsx(styles.arrow, styles.next)}
											aria-label="Следующий месяц"
											onClick={() => setMonth(shiftMonth(current, 1))}
										>
											<Icon name="chevron" size={17} />
										</button>
									</div>
								</div>
								<table className={styles.table}>
									<caption className="visually-hidden">Периоды планов по дням</caption>
									<thead>
										<tr>{WEEKDAYS.map((day) => <th key={day} scope="col">{day}</th>)}</tr>
									</thead>
									<tbody>
										{weeks.map((week) => (
											<tr key={week[0]?.date}>
												{week.map((cell, index) => (
													<DayCell
														key={cell.date}
														cell={cell}
														plan={planOn(cell.date)}
														isWeekStart={index === 0}
														isToday={cell.date === today}
													/>
												))}
											</tr>
										))}
									</tbody>
								</table>
								<div className={styles.legend}>
									<span><i />Период запланированного развития</span>
									<span role="status">{coveredDays > 0 ? `${coveredDays} дн. с планами` : 'В этом месяце планов нет'}</span>
								</div>
							</section>
							{selected && (
								<section className={clsx(styles.card, styles.selected)} aria-label="Выбранный план">
									<div className={styles.selectedHead}>
										<span className={styles.icon}>
											<Icon name={planIcon(selected)} size={20} />
										</span>
										{selectedState && (
											<Badge tone={selectedState.tone}>{selectedState.label}</Badge>
										)}
									</div>
									<h3>{selected.title}</h3>
									<p className={styles.period}>
										<Icon name="calendar" size={15} /> {formatPeriodLong(selected)}
									</p>
									<div className={styles.budget}>
										<small>{`Ориентир бюджета${goalName(selected) ? ` · ${goalName(selected)}` : ''}`}</small>
										<strong>{formatMoney(selected.budgetRub)}</strong>
									</div>
									{selected.note && <p className={styles.noteText}>{selected.note}</p>}
									<div className={styles.tracking}>
										{selected.dateMode === 'UNSCHEDULED' || !selected.startDate ? (
											<>
												<span>ПОДБОР ДОСТУПЕН СРАЗУ</span>
												<p>Подбор можно показывать сразу.</p>
											</>
										) : (
											<>
												<span>{`ПОДБОР НАЧНЁТСЯ ${formatPlanStart(selected).toUpperCase()}`}</span>
												<p>
													{selected.startDate > today
														? `Подбор начнётся с ${formatPlanStart(selected)}. Пока план впереди — на главной варианты ещё не появятся.`
														: 'План уже идёт: подходящие варианты могут появляться на главной.'}
												</p>
											</>
										)}
									</div>
									<div className={styles.notify}>
										{selected.notifications.inApp && <span>В кабинете</span>}
										{selected.notifications.email && <span>Письмо</span>}
									</div>
									<Link to={ROUTES.home} className={styles.homeLink}>
										Подбор на главной <Icon name="arrow" size={15} />
									</Link>
								</section>
							)}
						</div>
						<div className={styles.horizon}>
							<div>
								<h2>На горизонте</h2>
								<p>Переключайтесь между планами, чтобы посмотреть их период.</p>
							</div>
							<div className={styles.shortcuts}>
								{plans.map((plan) => {
									const isSelected = plan.id === selected?.id;

									return (
										<button
											key={plan.id}
											type="button"
											className={clsx(styles.shortcut, isSelected && styles.shortcutSelected)}
											aria-pressed={isSelected}
											aria-label={plan.title}
											onClick={() => choose(plan)}
										>
											<span className={styles.shortcutIcon}>
												<Icon name={planIcon(plan)} size={20} />
											</span>
											<span>
												<strong>{plan.title}</strong>
												<small>{shortcutHint(plan)}</small>
											</span>
											<Icon name="arrow" size={15} />
										</button>
									);
								})}
							</div>
						</div>
					</div>
				)}
				{tab === 'plans' && (
					<div role="tabpanel" id={panelId('plans')} aria-labelledby={tabId('plans')}>{cards(plans)}</div>
				)}
				{tab === 'ideas' && (
					<div role="tabpanel" id={panelId('ideas')} aria-labelledby={tabId('ideas')} className={styles.panel}>
						<div className={styles.explainer}>
							<span><Icon name="catalog" size={22} /></span>
							<div>
								<h2>Хорошей идее не обязательно иметь дату</h2>
								<p>
									Сохраните цель и примерный бюджет. Подбор по идее можно показывать на главной
									уже сейчас, а период добавить позже.
								</p>
							</div>
						</div>
						{ideas.length === 0 ? <p>Идей без даты пока нет.</p> : cards(ideas)}
					</div>
				)}
				<section className={styles.how} aria-label="Как это работает">
					{STEPS.map((step, index) => (
						<div key={step.title}>
							<span>{`0${index + 1}`}</span>
							<h3>{step.title}</h3>
							<p>{step.text}</p>
						</div>
					))}
				</section>
			</div>
		</AsyncBoundary>
	);
};
