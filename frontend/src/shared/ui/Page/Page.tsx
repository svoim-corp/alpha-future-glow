import { type ReactNode } from 'react';

import styles from './Page.module.scss';

export interface PageProps {
	title: string;
	subtitle?: string;
	actions?: ReactNode;
	children: ReactNode;
}

/** Каркас страницы: заголовок h1, пояснение, главные действия и содержимое */
export const Page = ({ title, subtitle, actions, children }: PageProps) => (
	<main className={styles.page}>
		<div className={styles.header}>
			<div>
				<h1 className={styles.title}>{title}</h1>
				{subtitle && <p className={styles.subtitle}>{subtitle}</p>}
			</div>
			{actions && <div className={styles.actions}>{actions}</div>}
		</div>
		{children}
	</main>
);
