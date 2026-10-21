import { type Goal } from '@entities/product';
import { Card, Field, RadioGroup } from '@shared/ui';

import { type CatalogInputErrors } from '../model/input';
import styles from './ProductCatalog.module.scss';

interface TaskPickerProps {
	goals: Goal[];
	goal: string | null;
	amountText: string;
	termText: string;
	errors: CatalogInputErrors;
	fitLine: string | null;
	onGoalChange: (goal: string) => void;
	onAmountChange: (text: string) => void;
	onTermChange: (text: string) => void;
}

/** «Подбор по задаче»: цель, сумма, срок → рекомендация */
export const TaskPicker = ({
	goals,
	goal,
	amountText,
	termText,
	errors,
	fitLine,
	onGoalChange,
	onAmountChange,
	onTermChange,
}: TaskPickerProps) => (
	<Card className={styles.picker} aria-labelledby="task-picker-title">
		<h2 id="task-picker-title" className={styles.pickerTitle}>Подбор по задаче</h2>
		<RadioGroup
			legend="На что нужны деньги"
			value={goal}
			onChange={onGoalChange}
			options={goals.map((item) => ({ value: item.code, label: item.name }))}
		/>
		<div className={styles.inputs}>
			<Field
				label="Сколько нужно, ₽"
				inputMode="numeric"
				value={amountText}
				error={errors.amount}
				onChange={(event) => onAmountChange(event.target.value)}
			/>
			<Field
				label="На какой срок, месяцев"
				inputMode="numeric"
				value={termText}
				error={errors.term}
				onChange={(event) => onTermChange(event.target.value)}
			/>
		</div>
		{fitLine && <p className={styles.fitLine} role="status">{fitLine}</p>}
	</Card>
);
