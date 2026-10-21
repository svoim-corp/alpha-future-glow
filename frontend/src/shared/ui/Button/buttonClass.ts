import clsx from 'clsx';

import styles from './Button.module.scss';

export type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'danger';

export const buttonClass = (variant: ButtonVariant, compact: boolean, className?: string) =>
	clsx(styles.button, styles[variant], compact && styles.compact, className);
