import { AsyncBoundary, Icon } from '@shared/ui';

import { useGetFaqQuery } from '../api/supportApi';
import { type FaqStep } from '../model/types';
import styles from './FaqList.module.scss';

interface FaqListProps {
	step: FaqStep;
	/** Поиск по вопросам и ответам: пустая строка показывает всё */
	query?: string;
}

export const FaqList = ({ step, query = '' }: FaqListProps) => {
	const { data, isLoading, isError, refetch } = useGetFaqQuery(step);
	const needle = query.trim().toLowerCase();
	const items = (data ?? []).filter((item) => (
		needle === '' || `${item.question} ${item.answer}`.toLowerCase().includes(needle)
	));

	return (
		<AsyncBoundary
			isLoading={isLoading}
			isError={isError}
			isEmpty={data?.length === 0}
			onRetry={() => void refetch()}
			loadingLabel="Загружаем вопросы"
			errorText="Не удалось загрузить вопросы"
			emptyTitle="По этому шагу вопросов пока нет"
		>
			<div className={styles.list}>
				{items.map((item) => (
					<details key={item.id} className={styles.item}>
						<summary>
							{item.question}
							<Icon name="chevron" size={16} />
						</summary>
						<p>{item.answer}</p>
					</details>
				))}
				{items.length === 0 && (
					<p className={styles.none} role="status">Подходящий ответ не найден. Напишите нам в форме рядом.</p>
				)}
			</div>
		</AsyncBoundary>
	);
};
