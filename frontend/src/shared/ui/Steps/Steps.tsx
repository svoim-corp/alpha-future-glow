import clsx from 'clsx';

import styles from './Steps.module.scss';

export interface StepsProps {
	labels: string[];
	/** Индекс текущего шага, с нуля */
	current: number;
	label: string;
}

/** Этап заполнения заявки; отделён от банковского статуса рассмотрения */
export const Steps = ({ labels, current, label }: StepsProps) => (
	<ol className={styles.steps} aria-label={label}>
		{labels.map((text, index) => (
			<li
				key={text}
				data-step={index + 1}
				className={clsx(styles.step, index === current && styles.current, index < current && styles.complete)}
				aria-current={index === current ? 'step' : undefined}
			>
				{text}
			</li>
		))}
	</ol>
);
