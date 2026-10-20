import { PaymentStatusBadge, type ScheduleItem, useGetScheduleQuery, useGetUpcomingPaymentsQuery } from '@entities/loan';
import { buildMonthGrid, monthRange, type MonthRef, shiftMonth } from '@shared/lib/calendar';
import { formatDate, formatMoney } from '@shared/lib/format';
import { AsyncBoundary, Button, Card } from '@shared/ui';
import { useState } from 'react';

import styles from './PaymentCalendar.module.scss';

const WEEKDAYS = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'];
const monthFormatter = new Intl.DateTimeFormat('ru-RU', { month: 'long', year: 'numeric' });

const startMonth = (): MonthRef => {
	const now = new Date();

	return { year: now.getFullYear(), month: now.getMonth() };
};

const Upcoming = () => {
	const { data, isLoading, isError, refetch } = useGetUpcomingPaymentsQuery(3);

	return (
		<AsyncBoundary
			isLoading={isLoading}
			isError={isError}
			isEmpty={data?.length === 0}
			onRetry={() => void refetch()}
			loadingLabel="Загружаем ближайшие платежи"
			errorText="Не удалось загрузить ближайшие платежи"
			emptyTitle="Ближайших платежей нет"
		>
			<ul className={styles.upcoming} aria-label="Ближайшие платежи">
				{data?.map((item) => (
					<li key={`${item.loanId}-${item.date}`}>
						{`${formatDate(item.date)} — ${formatMoney(item.amount)}, ${item.productName} `}
						<PaymentStatusBadge status={item.status} />
					</li>
				))}
			</ul>
		</AsyncBoundary>
	);
};

export const PaymentCalendar = () => {
	const [current, setCurrent] = useState<MonthRef>(startMonth);
	const { data, isLoading, isError, refetch } = useGetScheduleQuery(monthRange(current));
	const byDate = new Map<string, ScheduleItem[]>();

	for (const item of data ?? []) {
		byDate.set(item.date, [...(byDate.get(item.date) ?? []), item]);
	}

	const label = monthFormatter.format(new Date(current.year, current.month, 1));

	return (
		<div className={styles.wrap}>
			<Card aria-labelledby="cal-upcoming">
				<h2 id="cal-upcoming">Ближайшие платежи</h2>
				<Upcoming />
			</Card>
			<Card aria-labelledby="cal-month">
				<div className={styles.nav}>
					<Button variant="secondary" compact onClick={() => setCurrent((prev) => shiftMonth(prev, -1))}>
						Предыдущий месяц
					</Button>
					<h2 id="cal-month" className={styles.month}>{label}</h2>
					<Button variant="secondary" compact onClick={() => setCurrent((prev) => shiftMonth(prev, 1))}>
						Следующий месяц
					</Button>
				</div>
				<AsyncBoundary
					isLoading={isLoading}
					isError={isError}
					onRetry={() => void refetch()}
					loadingLabel="Загружаем платежи месяца"
					errorText="Не удалось загрузить платежи"
				>
					<div className={styles.scroll}>
						<table className={styles.table}>
							<caption className="visually-hidden">{`Платежи за ${label}`}</caption>
							<thead>
								<tr>{WEEKDAYS.map((day) => <th key={day} scope="col">{day}</th>)}</tr>
							</thead>
							<tbody>
								{buildMonthGrid(current).map((week) => (
									<tr key={week.find(Boolean) ?? 'empty'}>
										{week.map((date, index) => (
											<td key={date ?? `empty-${index}`} className={styles.cell}>
												{date && (
													<>
														<span className={styles.day}>{Number(date.slice(8))}</span>
														{byDate.get(date)?.map((item) => (
															<span key={item.loanId} className={styles.pay}>
																{`Платёж ${formatMoney(item.amount)}`}
															</span>
														))}
													</>
												)}
											</td>
										))}
									</tr>
								))}
							</tbody>
						</table>
					</div>
				</AsyncBoundary>
			</Card>
		</div>
	);
};
