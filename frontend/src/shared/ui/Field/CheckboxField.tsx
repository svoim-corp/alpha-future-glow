import { type InputHTMLAttributes, type ReactNode } from 'react';

import styles from './Field.module.scss';
import { FieldMessages } from './FieldMessages';
import { useFieldIds } from './useFieldIds';

export interface CheckboxFieldProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'type'> {
	label: ReactNode;
	error?: string;
}

/** Согласия — всегда отдельная галочка без предвыбора */
export const CheckboxField = ({ label, error, id, 'aria-describedby': describedBy, ...rest }: CheckboxFieldProps) => {
	const ids = useFieldIds({ id, error, describedBy });

	return (
		<div className={styles.field}>
			<label className={styles.choice} htmlFor={ids.fieldId}>
				<input
					{...rest}
					id={ids.fieldId}
					type="checkbox"
					aria-invalid={error ? true : undefined}
					aria-describedby={ids.describedBy}
				/>
				<span>{label}</span>
			</label>
			<FieldMessages error={error} errorId={ids.errorId} />
		</div>
	);
};
