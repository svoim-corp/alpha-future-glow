import { formatDateTime } from '@shared/lib/format';
import { AsyncBoundary, Badge, type Tone } from '@shared/ui';

import { useGetTicketsQuery } from '../api/supportApi';
import { type TicketStatus } from '../model/types';
import styles from './TicketList.module.scss';

const STATUS: Record<TicketStatus, { label: string; tone: Tone }> = {
	OPEN: { label: 'Рассматривается', tone: 'info' },
	ANSWERED: { label: 'Есть ответ', tone: 'success' },
	CLOSED: { label: 'Закрыто', tone: 'neutral' },
};

export const TicketList = () => {
	const { data, isLoading, isError, refetch } = useGetTicketsQuery();

	return (
		<AsyncBoundary
			isLoading={isLoading}
			isError={isError}
			isEmpty={data?.length === 0}
			onRetry={() => void refetch()}
			loadingLabel="Загружаем обращения"
			errorText="Не удалось загрузить обращения"
			emptyTitle="Обращений пока нет"
			emptyText="Если не нашли ответ в вопросах, напишите нам."
		>
			<ul className={styles.list} aria-label="Мои обращения">
				{data?.map((ticket) => (
					<li key={ticket.id}>
						<article className={styles.ticket} aria-label={ticket.subject}>
							<div className={styles.head}>
								<h3>{ticket.subject}</h3>
								<Badge tone={STATUS[ticket.status].tone}>{STATUS[ticket.status].label}</Badge>
							</div>
							<p className={styles.meta}>{`Обновлено ${formatDateTime(ticket.updatedAt)}`}</p>
							{ticket.reply && <p className={styles.reply}>{`Ответ: ${ticket.reply}`}</p>}
						</article>
					</li>
				))}
			</ul>
		</AsyncBoundary>
	);
};
