import { PaymentStatusBadge, type ScheduleItem, useGetScheduleQuery, useGetUpcomingPaymentsQuery } from '@entities/loan';
import { buildMonthGrid, monthRange, type MonthRef, shiftMonth } from '@shared/lib/calendar';
import { formatDate, formatMoney } from '@shared/lib/format';
import { AsyncBoundary, Button, Card, MonthPicker } from '@shared/ui';
import { useId, useState } from 'react';

import styles from './PaymentCalendar.module.scss';

const WEEKDAYS = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'];
const monthFormat = new Intl.DateTimeFormat('ru-RU', { month: 'long', year: 'numeric' });
/** «июнь 2027 г.» → «июнь 2027»: сокращение года в заголовке лишнее */
const monthFormatter = { format: (date: Date) => monthFormat.format(date).replace(/\s*г\.$/, '') };
const paymentPlural = new Intl.PluralRules('ru-RU');
const morePayments = (count: number) => {
	const form = paymentPlural.select(count);

	return `Ещё ${count} ${form === 'one' ? 'платёж' : form === 'few' ? 'платежа' : 'платежей'}`;
};

const startMonth = (): MonthRef => {
	const now = new Date();

	return { year: now.getUTCFullYear(), month: now.getUTCMonth() };
};

const Upcoming = () => {
	const { data, isLoading, isError, refetch } = useGetUpcomingPaymentsQuery(3, { refetchOnFocus: true, refetchOnReconnect: true, refetchOnMountOrArgChange: true });

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
					<li key={`${item.loanId}-${item.date}`} className={styles.item}>
						<span className={styles.itemDate}>{formatDate(item.date)}</span>
						<span className={styles.itemAmount}>{formatMoney(item.amount)}</span>
						<span className={styles.itemMeta}>
							{item.productName}
							<PaymentStatusBadge status={item.status} />
						</span>
					</li>
				))}
			</ul>
		</AsyncBoundary>
	);
};

export const PaymentCalendar = () => {
	const selectedId = useId();
	const today = new Date().toISOString().slice(0, 10);
	const [current, setCurrent] = useState<MonthRef>(startMonth);
	const [selected, setSelected] = useState(() => new Date().toISOString().slice(0, 10));
	const range = monthRange(current);
	const selectedDate = selected >= range.from && selected <= range.to ? selected : range.from;
	const { currentData: data, isLoading, isFetching, isError, refetch } = useGetScheduleQuery(range, { refetchOnFocus: true, refetchOnReconnect: true, refetchOnMountOrArgChange: true });
	const byDate = new Map<string, ScheduleItem[]>();

	for (const item of data ?? []) {
		byDate.set(item.date, [...(byDate.get(item.date) ?? []), item]);
	}

	const label = monthFormatter.format(new Date(current.year, current.month, 1));

	return (
		<div className={styles.wrap}>
			<Card aria-labelledby="cal-upcoming">
				<h2 id="cal-upcoming" className={styles.title}>Ближайшие платежи</h2>
				<Upcoming />
			</Card>
			<Card aria-labelledby="cal-month">
				<div className={styles.controls}>
					<MonthPicker value={current} onChange={setCurrent} />
					<Button variant="secondary" onClick={() => { setCurrent(startMonth()); setSelected(new Date().toISOString().slice(0, 10)); }}>Сегодня</Button>
				</div>
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
					isLoading={isLoading || (isFetching && !data)}
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
										{week.map((date, index) => {
											const payments = date ? byDate.get(date) ?? [] : [];
											const firstPayment = payments[0];

											return (
												<td key={date ?? `empty-${index}`} className={date ? styles.cell : `${styles.cell} ${styles.empty}`}>
													{date && (
														<button type="button" className={styles.day} aria-pressed={selectedDate === date}
															aria-controls={selectedId} aria-current={date === today ? 'date' : undefined}
															aria-label={`${formatDate(date)}, платежей: ${payments.length}. Показать платежи дня`}
															onClick={() => setSelected(date)}>
															<span className={styles.dayNumber}>{Number(date.slice(8))}</span>
															{firstPayment && <span className={styles.pay} data-status={firstPayment.status}>
																<span className={styles.payAmount}>{formatMoney(firstPayment.amount)}</span>
																<span className={styles.payProduct}>{firstPayment.productName}</span>
															</span>}
															{payments.length > 1 && <span className={styles.more} title={morePayments(payments.length - 1)}>
																<span className={styles.moreFull}>{morePayments(payments.length - 1)}</span>
																<span className={styles.moreCompact} aria-hidden="true">+{payments.length - 1}</span>
															</span>}
														</button>
													)}
												</td>
											);
										})}
									</tr>
								))}
							</tbody>
						</table>
					</div>
					<section id={selectedId} className={styles.selected} aria-label="Платежи выбранного дня" aria-live="polite">
						<h3>Платежи {formatDate(selectedDate)}</h3>
						{(byDate.get(selectedDate) ?? []).length > 0 ? <ul className={styles.upcoming}>
							{byDate.get(selectedDate)?.map((item) => <li className={styles.item} key={item.loanId}>
								<strong>{item.productName}</strong><span>{formatMoney(item.amount)}</span><PaymentStatusBadge status={item.status} />
							</li>)}
						</ul> : <p>На эту дату платежей нет.</p>}
					</section>
				</AsyncBoundary>
			</Card>
		</div>
	);
};
