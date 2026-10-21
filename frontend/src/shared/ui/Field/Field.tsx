import clsx from 'clsx';
import { type InputHTMLAttributes, type ReactNode } from 'react';

import styles from './Field.module.scss';
import { FieldMessages } from './FieldMessages';
import { useFieldIds } from './useFieldIds';

export interface FieldProps extends InputHTMLAttributes<HTMLInputElement> {
	label: ReactNode;
	hint?: string;
	error?: string;
}

export const Field = ({ label, hint, error, id, className, 'aria-describedby': describedBy, ...rest }: FieldProps) => {
	const ids = useFieldIds({ id, hint, error, describedBy });

	return (
		<div className={styles.field}>
			<label className={styles.label} htmlFor={ids.fieldId}>{label}</label>
			<input
				{...rest}
				id={ids.fieldId}
				className={clsx(styles.control, className)}
				aria-invalid={error ? true : undefined}
				aria-describedby={ids.describedBy}
			/>
			<FieldMessages hint={hint} hintId={ids.hintId} error={error} errorId={ids.errorId} />
		</div>
	);
};
