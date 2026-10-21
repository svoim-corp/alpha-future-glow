import clsx from 'clsx';
import { type ReactNode, type SelectHTMLAttributes } from 'react';

import styles from './Field.module.scss';
import { FieldMessages } from './FieldMessages';
import { useFieldIds } from './useFieldIds';

export interface SelectOption {
	value: string;
	label: string;
}

export interface SelectFieldProps extends SelectHTMLAttributes<HTMLSelectElement> {
	label: ReactNode;
	options: SelectOption[];
	hint?: string;
	error?: string;
}

export const SelectField = ({
	label,
	options,
	hint,
	error,
	id,
	className,
	'aria-describedby': describedBy,
	...rest
}: SelectFieldProps) => {
	const ids = useFieldIds({ id, hint, error, describedBy });

	return (
		<div className={styles.field}>
			<label className={styles.label} htmlFor={ids.fieldId}>{label}</label>
			<select
				{...rest}
				id={ids.fieldId}
				className={clsx(styles.control, className)}
				aria-invalid={error ? true : undefined}
				aria-describedby={ids.describedBy}
			>
				{options.map((option) => (
					<option key={option.value} value={option.value}>{option.label}</option>
				))}
			</select>
			<FieldMessages hint={hint} hintId={ids.hintId} error={error} errorId={ids.errorId} />
		</div>
	);
};
