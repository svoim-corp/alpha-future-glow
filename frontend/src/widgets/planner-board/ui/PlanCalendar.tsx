import { type BusinessPlan, plansOnDate } from '@entities/plan';
import { buildMonthCells, type MonthRef } from '@shared/lib/calendar';
import { Icon, MonthPicker } from '@shared/ui';
import clsx from 'clsx';

import styles from './PlannerBoard.module.scss';

interface PlanCalendarProps {
	month: MonthRef;
	plans: BusinessPlan[];
	today: string;
	selectedDate: string | null;
	onSelect: (date: string) => void;
	onMove: (delta: number) => void;
	onToday: () => void;
	onMonthChange: (month: MonthRef) => void;
}

const weekdays = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'];
const dateFormat = new Intl.DateTimeFormat('ru-RU', { day: 'numeric', month: 'long', year: 'numeric', timeZone: 'UTC' });
export const calendarDateLabel = (date: string): string => dateFormat.format(new Date(`${date}T00:00:00Z`)).replace(/\s*г\.$/u, '');
const monthFormat = new Intl.DateTimeFormat('ru-RU', { month: 'long', year: 'numeric', timeZone: 'UTC' });

export const PlanCalendar = ({
	month, plans, today, selectedDate, onSelect, onMove, onToday, onMonthChange,
}: PlanCalendarProps) => {
	const title = monthFormat.format(new Date(Date.UTC(month.year, month.month, 1))).replace(/\s*г\.$/u, '');

	return (
		<section className={styles.card} aria-label="Календарь планов">
			<div className={styles.calendarHead}>
				<div><span className={styles.label}>КАЛЕНДАРЬ РАЗВИТИЯ</span><h2 className={styles.month} aria-live="polite">{title.charAt(0).toUpperCase() + title.slice(1)}</h2></div>
				<div className={styles.controls}>
					<button className={styles.today} type="button" onClick={onToday}>Сегодня</button>
					<button className={clsx(styles.arrow, styles.prev)} type="button" aria-label="Предыдущий месяц" onClick={() => onMove(-1)}><Icon name="chevron" size={16} /></button>
					<button className={clsx(styles.arrow, styles.next)} type="button" aria-label="Следующий месяц" onClick={() => onMove(1)}><Icon name="chevron" size={16} /></button>
				</div>
			</div>
			<MonthPicker value={month} onChange={onMonthChange} />
			<table className={styles.table}>
				<caption className="visually-hidden">Планы развития по дням. Выберите дату, чтобы увидеть все её планы.</caption>
				<thead><tr>{weekdays.map((day) => <th key={day} scope="col">{day}</th>)}</tr></thead>
				<tbody>{buildMonthCells(month).map((week) => (
					<tr key={week[0]?.date}>{week.map((cell) => {
						const items = plansOnDate(plans, cell.date);

						return (
							<td key={cell.date}>
								<button
									type="button"
									className={clsx(
										styles.day, cell.outside && styles.outside, items.length > 0 && styles.covered,
									)}
									aria-label={`${calendarDateLabel(cell.date)}, планов: ${items.length}`}
									aria-current={cell.date === today ? 'date' : undefined}
									aria-pressed={selectedDate === cell.date}
									onClick={() => onSelect(cell.date)}
								>
									<span className={clsx(styles.dayLabel, cell.date === today && styles.todayMark)}>
										{Number(cell.date.slice(8))}
									</span>
									{items.length > 0 && <span className={styles.event} aria-hidden="true"><span className={styles.eventTitle}>{items[0]?.title}</span><span className={styles.count}>{items.length}</span></span>}
								</button>
							</td>
						);
					})}</tr>
				))}</tbody>
			</table>
			<p className={styles.legend}>Выберите день, чтобы увидеть все планы. Сегодня — по московскому времени.</p>
		</section>
	);
};
