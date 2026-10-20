import { type ButtonHTMLAttributes } from 'react';

import { Spinner } from '../Spinner';
import { buttonClass, type ButtonVariant } from './buttonClass';

export type { ButtonVariant };

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
	variant?: ButtonVariant;
	loading?: boolean;
	compact?: boolean;
}

export const Button = ({
	variant = 'primary',
	loading = false,
	compact = false,
	disabled,
	className,
	children,
	type = 'button',
	...rest
}: ButtonProps) => (
	<button
		{...rest}
		type={type}
		className={buttonClass(variant, compact, className)}
		disabled={disabled || loading}
		aria-busy={loading}
	>
		{loading && <Spinner />}
		{children}
	</button>
);
