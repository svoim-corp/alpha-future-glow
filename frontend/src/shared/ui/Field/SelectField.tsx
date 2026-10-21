import { type ReactNode } from 'react';

import { Select, type SelectOption } from '../Select';
import styles from './Field.module.scss';
import { FieldMessages } from './FieldMessages';
import { useFieldIds } from './useFieldIds';

export interface SelectFieldProps {
	label: ReactNode;
	options: SelectOption[];
	value: string;
	onChange: (value: string) => void;
	hint?: string;
	error?: string;
	id?: string;
	name?: string;
	disabled?: boolean;
	className?: string;
	'aria-describedby'?: string;
}

export const SelectField = ({
	label,
	options,
	value,
	onChange,
	hint,
	error,
	id,
	name,
	disabled,
	className,
	'aria-describedby': describedBy,
}: SelectFieldProps) => {
	const ids = useFieldIds({ id, hint, error, describedBy });

	return (
		<div className={styles.field}>
			<label className={styles.label} htmlFor={ids.fieldId}>{label}</label>
			<Select
				id={ids.fieldId}
				name={name}
				className={className}
				options={options}
				value={value}
				onChange={onChange}
				disabled={disabled}
				invalid={Boolean(error)}
				aria-describedby={ids.describedBy}
			/>
			<FieldMessages hint={hint} hintId={ids.hintId} error={error} errorId={ids.errorId} />
		</div>
	);
};
