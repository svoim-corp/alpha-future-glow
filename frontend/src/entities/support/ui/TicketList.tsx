import { formatDateTime } from '@shared/lib/format';
import { AsyncBoundary, Badge, Card, type Tone } from '@shared/ui';

import { useGetTicketsQuery } from '../api/supportApi';
import { type TicketStatus } from '../model/types';

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
			emptyText="Если не нашли ответ в вопросах выше, напишите нам."
		>
			<ul aria-label="Мои обращения">
				{data?.map((ticket) => (
					<li key={ticket.id}>
						<Card as="article" aria-label={ticket.subject}>
							<h3>{ticket.subject}</h3>
							<Badge tone={STATUS[ticket.status].tone}>{STATUS[ticket.status].label}</Badge>
							<p>{`Обновлено ${formatDateTime(ticket.updatedAt)}`}</p>
							{ticket.reply && <p>{`Ответ: ${ticket.reply}`}</p>}
						</Card>
					</li>
				))}
			</ul>
		</AsyncBoundary>
	);
};
