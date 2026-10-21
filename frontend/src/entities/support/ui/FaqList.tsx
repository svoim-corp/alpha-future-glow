import { AsyncBoundary } from '@shared/ui';

import { useGetFaqQuery } from '../api/supportApi';
import { type FaqStep } from '../model/types';

export const FaqList = ({ step }: { step: FaqStep }) => {
	const { data, isLoading, isError, refetch } = useGetFaqQuery(step);

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
			{data?.map((item) => (
				<details key={item.id}>
					<summary>{item.question}</summary>
					<p>{item.answer}</p>
				</details>
			))}
		</AsyncBoundary>
	);
};
