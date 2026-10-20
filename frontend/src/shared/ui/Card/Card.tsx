import clsx from 'clsx';
import { type ElementType, type HTMLAttributes } from 'react';

import styles from './Card.module.scss';

export interface CardProps extends HTMLAttributes<HTMLElement> {
	accent?: boolean;
	as?: ElementType;
}

export const Card = ({ accent = false, as: Tag = 'section', className, ...rest }: CardProps) => (
	<Tag {...rest} className={clsx(styles.card, accent && styles.accent, className)} />
);
