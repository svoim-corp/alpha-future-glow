import '@daypicker/react/style.css';

import { DayPicker } from '@daypicker/react';
import { ru } from '@daypicker/react/locale';
import { type KeyboardEvent, useEffect, useId, useRef, useState } from 'react';

import { Field, type FieldProps } from '../Field';
import { Icon } from '../Icon';
import { Modal } from '../Modal';
import { MonthPicker } from '../MonthPicker';
import styles from './DateField.module.scss';
import { calendarIso, displayCalendarDate, parseCalendarDate } from './dates';

export interface DateFieldProps extends Omit<FieldProps, 'value' | 'onChange' | 'type' | 'placeholder'> {
	/** ISO YYYY-MM-DD; незавершённый ввод остаётся строкой для валидации формы. */
	value: string;
	onValueChange: (value: string) => void;
}

export const DateField = ({ value, onValueChange, disabled, id, ...props }: DateFieldProps) => {
	const generated = useId();
	const fieldId = id ?? generated;
	const root = useRef<HTMLDivElement>(null);
	const panel = useRef<HTMLDivElement>(null);
	const trigger = useRef<HTMLButtonElement>(null);
	const [open, setOpen] = useState(false);
	const [month, setMonth] = useState(() => parseCalendarDate(value) ?? new Date());
	const selected = parseCalendarDate(value);
	const close = () => { setOpen(false); trigger.current?.focus(); };
	const escape = (event: KeyboardEvent<HTMLDivElement>) => {
		if (!open || event.key !== 'Escape') return;
		if (event.target instanceof Element && event.target.closest('[role="combobox"][aria-expanded="true"]')) return;
		event.preventDefault();
		event.stopPropagation();
		close();
	};

	useEffect(() => {
		if (!open || disabled) return;
		const day = panel.current?.querySelector<HTMLButtonElement>('.rdp-selected button')
			?? panel.current?.querySelector<HTMLButtonElement>('.rdp-today button')
			?? panel.current?.querySelector<HTMLButtonElement>('.rdp-day_button');
		day?.focus();
		const outside = (event: PointerEvent) => {
			const dialog = panel.current?.closest('[role="dialog"]');
			if (event.target instanceof Node
				&& !root.current?.contains(event.target) && !dialog?.contains(event.target)) setOpen(false);
		};
		document.addEventListener('pointerdown', outside);
		return () => document.removeEventListener('pointerdown', outside);
	}, [open, disabled]);

	return (
		<div ref={root} className={styles.root} onKeyDownCapture={escape}>
			<div className={styles.input}>
				<Field {...props} id={fieldId} disabled={disabled} type="text" value={displayCalendarDate(value)}
					placeholder="ДД.ММ.ГГГГ" inputMode="numeric" autoComplete="off"
					onChange={(event) => {
						const text = event.target.value;
						const date = parseCalendarDate(text);
						onValueChange(date ? calendarIso(date) : text);
					}} />
				<button ref={trigger} className={styles.trigger} type="button" disabled={disabled}
					aria-label={`Выбрать дату: ${typeof props.label === 'string' ? props.label : 'календарь'}`}
					aria-haspopup="dialog"
					aria-expanded={open} aria-controls={open ? `${fieldId}-calendar` : undefined}
					onClick={() => {
						if (!open) setMonth(selected ?? new Date());
						setOpen(!open);
					}}><Icon name="calendar" size={20} /></button>
			</div>
			{open && !disabled && <Modal size="small" layer="overlay" onClose={close}
				title={`Выбор даты: ${typeof props.label === 'string' ? props.label : 'календарь'}`}>
				<div ref={panel} id={`${fieldId}-calendar`} className={styles.panel} role="group" aria-label="Выбор даты">
					<MonthPicker value={{ year: month.getFullYear(), month: month.getMonth() }}
						onChange={(next) => setMonth(new Date(next.year, next.month, 1))} />
					<DayPicker mode="single" required locale={ru} weekStartsOn={1} className={styles.calendar}
						month={month} onMonthChange={setMonth} selected={selected} showOutsideDays hideNavigation
						onSelect={(date) => {
							if (!date) return;
							onValueChange(calendarIso(date));
							close();
						}} />
				</div>
			</Modal>}
		</div>
	);
};
