import { type Range } from '@entities/product';
import { type CSSProperties, useId } from 'react';

import { fillPercent } from '../model/range';
import styles from './RangeField.module.scss';

interface RangeFieldProps {
	label: string;
	/** Значение, показанное справа от подписи: «1 500 000 ₽» */
	display: string;
	value: number;
	range: Range;
	step: number;
	minLabel: string;
	maxLabel: string;
	onChange: (value: number) => void;
}

/** Ползунок с подписью, текущим значением и границами (калькулятор макета catalog.html) */
export const RangeField = ({ label, display, value, range, step, minLabel, maxLabel, onChange }: RangeFieldProps) => {
	const id = useId();

	return (
		<div className={styles.field}>
			<div className={styles.head}>
				<label htmlFor={id}>{label}</label>
				<output htmlFor={id}>{display}</output>
			</div>
			<input
				id={id}
				type="range"
				min={range.min}
				max={range.max}
				step={step}
				value={value}
				style={{ '--fill': `${fillPercent(value, range)}%` } as CSSProperties}
				onChange={(event) => onChange(Number(event.target.value))}
			/>
			<div className={styles.bounds}><span>{minLabel}</span><span>{maxLabel}</span></div>
		</div>
	);
};
