import { type ReactNode, useId } from 'react';

import styles from './Field.module.scss';
import { FieldMessages } from './FieldMessages';

export interface RadioOption<T extends string> {
	value: T;
	label: string;
	/** Пояснение под названием, только для variant="cards" */
	description?: string;
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
	/** pills — компактные «таблетки»; cards — крупные карточки с точкой выбора и пояснением */
	variant?: 'pills' | 'cards';
}

/** Взаимоисключающие варианты; нативные radio остаются доступными с клавиатуры */
export const RadioGroup = <T extends string>({
	legend,
	name,
	options,
	value,
	onChange,
	error,
	hint,
	variant = 'pills',
}: RadioGroupProps<T>) => {
	const generated = useId();
	const groupName = name ?? generated;

	return (
		<fieldset className={styles.group}>
			<legend>{legend}</legend>
			<div className={variant === 'cards' ? styles.cards : styles.options}>
				{options.map((option) => {
					const titleId = `${groupName}-${option.value}-title`;
					const textId = `${groupName}-${option.value}-text`;

					if (variant === 'cards') {
						return (
							<label key={option.value} className={styles.card}>
								<input
									type="radio"
									name={groupName}
									value={option.value}
									checked={value === option.value}
									disabled={option.disabled}
									aria-labelledby={titleId}
									aria-describedby={option.description ? textId : undefined}
									onChange={() => onChange(option.value)}
								/>
								<span className={styles.dot} aria-hidden="true" />
								<span id={titleId} className={styles.cardTitle}>{option.label}</span>
								{option.description && (
									<span id={textId} className={styles.cardText}>{option.description}</span>
								)}
							</label>
						);
					}

					return (
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
					);
				})}
			</div>
			<FieldMessages hint={hint} error={error} />
		</fieldset>
	);
};
