import clsx from 'clsx';
import { type HTMLAttributes } from 'react';

import styles from './Badge.module.scss';

export type Tone = 'neutral' | 'info' | 'warning' | 'success' | 'danger';

export interface BadgeProps extends HTMLAttributes<HTMLSpanElement> {
	tone?: Tone;
}

export const Badge = ({ tone = 'neutral', className, ...rest }: BadgeProps) => (
	<span {...rest} className={clsx(styles.badge, tone !== 'neutral' && styles[tone], className)} />
);
