import { type ReactNode, useId } from 'react';

import styles from './Field.module.scss';
import { FieldMessages } from './FieldMessages';

export interface RadioOption<T extends string> {
	value: T;
	label: string;
	disabled?: boolean;
}

export interface RadioGroupProps<T extends string> {
	legend: ReactNode;
	name?: string;
	options: RadioOption<T>[];
	value: T | null;
	onChange: (value: T) => void;
	error?: string;
	hint?: string;
}

/** Взаимоисключающие варианты в виде «таблеток»; нативные radio остаются доступными с клавиатуры */
export const RadioGroup = <T extends string>({
	legend,
	name,
	options,
	value,
	onChange,
	error,
	hint,
}: RadioGroupProps<T>) => {
	const generated = useId();
	const groupName = name ?? generated;

	return (
		<fieldset className={styles.group}>
			<legend>{legend}</legend>
			<div className={styles.options}>
				{options.map((option) => (
					<label key={option.value} className={styles.option}>
						<input
							type="radio"
							name={groupName}
							value={option.value}
							checked={value === option.value}
							disabled={option.disabled}
							onChange={() => onChange(option.value)}
						/>
						<span>{option.label}</span>
					</label>
				))}
			</div>
			<FieldMessages hint={hint} error={error} />
		</fieldset>
	);
};
