import clsx from 'clsx';
import { type ReactNode } from 'react';

import styles from './Alert.module.scss';

export type AlertTone = 'info' | 'warning' | 'success' | 'danger';

export interface AlertProps {
	tone?: AlertTone;
	title?: string;
	children?: ReactNode;
	actions?: ReactNode;
	className?: string;
}

/** Ошибки и предупреждения объявляются скринридеру (role=alert), остальное — role=status */
export const Alert = ({ tone = 'info', title, children, actions, className }: AlertProps) => (
	<div
		className={clsx(styles.alert, tone !== 'info' && styles[tone], className)}
		role={tone === 'danger' || tone === 'warning' ? 'alert' : 'status'}
	>
		{title && <strong className={styles.title}>{title}</strong>}
		{children && <div className={styles.body}>{children}</div>}
		{actions && <div className={styles.actions}>{actions}</div>}
	</div>
);
