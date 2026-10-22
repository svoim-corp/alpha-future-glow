import { type MonthRef } from '../../lib/calendar';
import { SelectField } from '../Field';
import styles from './MonthPicker.module.scss';

export interface MonthPickerProps {
	value: MonthRef;
	onChange: (value: MonthRef) => void;
}

const months = Array.from({ length: 12 }, (_, month) => ({
	value: String(month),
	label: new Intl.DateTimeFormat('ru-RU', { month: 'long' }).format(new Date(2026, month, 1)),
}));

export const MonthPicker = ({ value, onChange }: MonthPickerProps) => {
	const firstYear = Math.min(1900, value.year);
	const lastYear = Math.max(new Date().getFullYear() + 100, value.year);
	const years = Array.from({ length: lastYear - firstYear + 1 }, (_, index) => {
		const year = String(firstYear + index);

		return { value: year, label: year };
	});

	return (
		<div className={styles.picker}>
			<SelectField label="Месяц" value={String(value.month)} options={months}
				onChange={(month) => onChange({ ...value, month: Number(month) })} />
			<SelectField label="Год" value={String(value.year)} options={years}
				onChange={(year) => onChange({ ...value, year: Number(year) })} />
		</div>
	);
};
