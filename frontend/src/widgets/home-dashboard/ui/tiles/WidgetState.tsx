import { Button } from '@shared/ui';
import { type ReactNode } from 'react';

import styles from '../Dashboard.module.scss';

interface WidgetStateProps {
	loading: boolean; error: boolean; retry: () => void; empty?: boolean; emptyText?: string; children: ReactNode;
}
export const WidgetState = ({ loading, error, retry, empty = false, emptyText = 'Пока нет данных.', children }: WidgetStateProps) => {
	if (loading) return <div className={styles.skeleton} role="status" aria-label="Загрузка виджета"><span /><span /><span /></div>;
	if (error) return <div className={styles.state} role="alert"><strong>Данные недоступны</strong><Button variant="secondary" compact onClick={retry}>Повторить</Button></div>;
	if (empty) return <p className={styles.empty}>{emptyText}</p>;
	return children;
};
