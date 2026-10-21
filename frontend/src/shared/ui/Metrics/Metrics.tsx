import clsx from 'clsx';
import { type ReactNode } from 'react';

import styles from './Metrics.module.scss';

export interface MetricItem {
	label: string;
	value: ReactNode;
	hint?: string;
	/** Рубиновая подсветка плитки: требует внимания */
	accent?: boolean;
}

export interface MetricsProps {
	items: MetricItem[];
	/** Подпись группы для скринридеров */
	label?: string;
}

/** Обзорные плитки над списком (макеты «Мои заявки», «Мои кредиты») */
export const Metrics = ({ items, label }: MetricsProps) => (
	<dl className={styles.grid} aria-label={label}>
		{items.map((item) => (
			<div key={item.label} className={clsx(styles.metric, item.accent && styles.accent)}>
				<dt>{item.label}</dt>
				<dd className={styles.value}>{item.value}</dd>
				{item.hint && <dd className={styles.hint}>{item.hint}</dd>}
			</div>
		))}
	</dl>
);
