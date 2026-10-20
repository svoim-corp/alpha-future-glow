import { Link, type LinkProps } from 'react-router';

import { buttonClass, type ButtonVariant } from './buttonClass';

export interface ButtonLinkProps extends LinkProps {
	variant?: ButtonVariant;
	compact?: boolean;
}

/** Ссылка, выглядящая как кнопка: переход по маршруту — это ссылка, а не button */
export const ButtonLink = ({ variant = 'primary', compact = false, className, ...rest }: ButtonLinkProps) => (
	<Link {...rest} className={buttonClass(variant, compact, className)} />
);
