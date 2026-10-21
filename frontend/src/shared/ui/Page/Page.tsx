import { type ReactNode } from 'react';

import styles from './Page.module.scss';

export interface PageProps {
	title: string;
	titleAside?: ReactNode;
	/** Строка над заголовком, например «КРЕДИТНЫЙ КАБИНЕТ · СЕВЕРНАЯ ПЕКАРНЯ» */
	eyebrow?: string;
	subtitle?: string;
	actions?: ReactNode;
	children: ReactNode;
}

/** Каркас страницы по макетам: eyebrow, заголовок h1, пояснение, главные действия и содержимое */
export const Page = ({ title, titleAside, eyebrow, subtitle, actions, children }: PageProps) => (
	<main className={styles.page}>
		<div className={styles.header}>
			<div>
				{eyebrow && <div className={styles.eyebrow}>{eyebrow}</div>}
				<div className={styles.titleRow}><h1 className={styles.title}>{title}</h1>{titleAside}</div>
				{subtitle && <p className={styles.subtitle}>{subtitle}</p>}
			</div>
			{actions && <div className={styles.actions}>{actions}</div>}
		</div>
		{children}
	</main>
);
