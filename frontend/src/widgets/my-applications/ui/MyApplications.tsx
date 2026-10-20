import { ApplicationCard, useGetMyApplicationsQuery } from '@entities/application';
import { DeleteDraftButton } from '@features/delete-draft';
import { ROUTES } from '@shared/config';
import { AsyncBoundary, ButtonLink } from '@shared/ui';

import styles from './MyApplications.module.scss';

/** Блок «Мои заявки»: черновики и поданные заявки компании, новые сверху */
export const MyApplications = () => {
	const { data, isLoading, isError, refetch } = useGetMyApplicationsQuery();
	const items = data?.items ?? [];

	return (
		<AsyncBoundary
			isLoading={isLoading}
			isError={isError}
			isEmpty={items.length === 0}
			onRetry={() => void refetch()}
			loadingLabel="Загрузка заявок"
			errorText="Не удалось загрузить заявки."
			emptyTitle="Заявок пока нет"
			emptyText="Подберите продукт в каталоге — черновик сохранится, и вернуться к нему можно в любой момент."
			emptyAction={<ButtonLink to={ROUTES.products}>Получить финансирование</ButtonLink>}
		>
			<ul className={styles.list} aria-label="Мои заявки">
				{items.map((application) => (
					<li key={application.id}>
						<ApplicationCard
							application={application}
							actions={application.status === 'DRAFT' ? (
								<DeleteDraftButton
									applicationId={application.id}
									productName={application.product.name}
								/>
							) : undefined}
						/>
					</li>
				))}
			</ul>
		</AsyncBoundary>
	);
};
