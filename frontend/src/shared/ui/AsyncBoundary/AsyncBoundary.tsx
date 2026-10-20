import { type ReactNode } from 'react';

import { Alert } from '../Alert';
import { Button } from '../Button';
import styles from './AsyncBoundary.module.scss';

export interface AsyncBoundaryProps {
	isLoading: boolean;
	isError: boolean;
	isEmpty?: boolean;
	onRetry: () => void;
	/** Подпись состояния загрузки для скринридера */
	loadingLabel?: string;
	errorText?: string;
	emptyTitle?: string;
	emptyText?: string;
	emptyAction?: ReactNode;
	children: ReactNode;
}

/**
 * Четыре состояния экрана с данными: загрузка, ошибка с повтором, пусто, данные.
 * При повторной загрузке известные данные не прячем.
 */
export const AsyncBoundary = ({
	isLoading,
	isError,
	isEmpty = false,
	onRetry,
	loadingLabel = 'Загрузка',
	errorText = 'Не удалось загрузить данные.',
	emptyTitle = 'Здесь пока пусто',
	emptyText,
	emptyAction,
	children,
}: AsyncBoundaryProps) => {
	if (isLoading) {
		return (
			<div className={styles.skeleton} role="status" aria-label={loadingLabel}>
				<div className={styles.line} />
				<div className={styles.line} />
				<div className={styles.line} />
			</div>
		);
	}

	if (isError) {
		return (
			<Alert
				tone="danger"
				title={errorText}
				actions={<Button variant="secondary" compact onClick={onRetry}>Повторить</Button>}
			/>
		);
	}

	if (isEmpty) {
		return (
			<div className={styles.empty}>
				<h3>{emptyTitle}</h3>
				{emptyText && <p>{emptyText}</p>}
				{emptyAction}
			</div>
		);
	}

	return children;
};
