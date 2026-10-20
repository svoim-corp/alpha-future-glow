import clsx from 'clsx';
import { type ButtonHTMLAttributes } from 'react';

import styles from './Button.module.scss';

export type ButtonVariant = 'primary' | 'secondary';

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
	variant?: ButtonVariant;
	loading?: boolean;
}

export const Button = ({
	variant = 'primary',
	loading = false,
	disabled,
	className,
	children,
	type = 'button',
	...rest
}: ButtonProps) => (
	<button
		{...rest}
		type={type}
		className={clsx(styles.button, styles[variant], className)}
		disabled={disabled || loading}
		aria-busy={loading}
	>
		{loading ? 'Загрузка…' : children}
	</button>
);
