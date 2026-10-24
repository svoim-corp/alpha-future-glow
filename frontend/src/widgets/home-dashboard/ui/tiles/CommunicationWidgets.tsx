import { type DashboardWidget } from '@entities/dashboard';
import { type Notification, useGetNotificationsQuery, useMarkNotificationReadMutation } from '@entities/notification';
import { type Ticket, type TicketStatus, useGetTicketsQuery } from '@entities/support';
import { getErrorMessage } from '@shared/api';
import { paths, ROUTES } from '@shared/config';
import { formatDateTime } from '@shared/lib/format';
import { Alert, Badge, Modal } from '@shared/ui';
import { useState } from 'react';
import { Link, useNavigate } from 'react-router';

import { notificationRoute } from '../../model/notificationRoute';
import styles from '../Dashboard.module.scss';
import { WidgetPreview } from './WidgetPreview';
import { WidgetState } from './WidgetState';

const freshness = { refetchOnFocus: true, refetchOnReconnect: true, refetchOnMountOrArgChange: true };
const statusLabels: Record<TicketStatus, string> = { OPEN: 'Открыто', ANSWERED: 'Есть ответ', CLOSED: 'Закрыто' };

export const NotificationsWidget = ({ widget }: { widget: Extract<DashboardWidget, { type: 'notifications' }> }) => {
	const query = useGetNotificationsQuery(widget.settings.filter === 'unread' ? { unread: true } : undefined, { ...freshness, pollingInterval: 60000 });
	const [markRead, mutation] = useMarkNotificationReadMutation();
	const [error, setError] = useState('');
	const [pendingRoute, setPendingRoute] = useState<string | null>(null);
	const [selected, setSelected] = useState<Notification | null>(null);
	const navigate = useNavigate();
	const items = query.currentData?.items ?? [];
	const open = async (item: Notification) => {
		const to = notificationRoute(item.link);
		setError(''); setPendingRoute(null);
		if (!item.read) {
			try { await markRead(item.id).unwrap(); }
			catch (failure) {
				setError(getErrorMessage(failure, 'Не удалось отметить уведомление прочитанным.'));
				setPendingRoute(to);
				if (!to) setSelected(item);
				return;
			}
		}
		if (to) void navigate(to);
		else setSelected(item);
	};
	return <><WidgetState loading={query.isLoading || (query.isFetching && !query.currentData)} error={query.isError} retry={() => void query.refetch()}>
		<p className={styles.muted}>Непрочитанных: {query.currentData?.unreadCount ?? '—'}</p>
		{error && <Alert tone="danger" title={error}>{pendingRoute && <Link to={pendingRoute}>Перейти к событию</Link>}</Alert>}
		{items.length === 0 ? <p className={styles.empty}>{widget.settings.filter === 'unread' ? 'Непрочитанных уведомлений нет.' : 'Событий пока нет.'}</p>
			: <WidgetPreview rowHeight={132} maxColumns={3} minColumnWidth={210}>{items.map((item) => <button key={item.id} type="button" className={styles.feedButton} data-unread={!item.read} disabled={mutation.isLoading} onClick={() => void open(item)}>
				<strong>{item.title}</strong><span className={styles.clamped}>{item.text}</span><small className={styles.muted}>{formatDateTime(item.createdAt)}{!item.read ? ' · Не прочитано' : ''}</small>
			</button>)}</WidgetPreview>}
		<p className={styles.muted}>Последние события{items.length >= 50 ? ' · выборка ограничена 50 уведомлениями' : ''}. Другие полученные уведомления — в колокольчике.</p>
	</WidgetState>
		{selected && <Modal title={selected.title} onClose={() => setSelected(null)}>
			{error && <Alert tone="danger" title={error} />}
			<p className={styles.muted}>{formatDateTime(selected.createdAt)}</p>
			<p className={styles.longText}>{selected.text}</p>
		</Modal>}
	</>;
};

export const TicketsWidget = ({ widget }: { widget: Extract<DashboardWidget, { type: 'tickets' }> }) => {
	const query = useGetTicketsQuery(undefined, freshness);
	const [selectedId, setSelectedId] = useState<string | null>(null);
	const items = [...(query.currentData ?? [])].filter((item) => widget.settings.filter === 'answered' ? item.status === 'ANSWERED'
		: widget.settings.filter === 'active' ? item.status !== 'CLOSED' : true).sort((a, b) => b.updatedAt.localeCompare(a.updatedAt));
	const selected: Ticket | undefined = query.currentData?.find((item) => item.id === selectedId);
	return <>
		<WidgetState loading={query.isLoading} error={query.isError} retry={() => void query.refetch()}>
			{(query.currentData?.length ?? 0) >= 100 && <p className={styles.muted} title="Показаны последние 100 обращений. Фильтр применяется к полученной выборке.">Последние 100 обращений</p>}
			{items.length === 0 ? <p className={styles.empty}>В полученном списке подходящих обращений нет.</p>
				: <WidgetPreview rowHeight={92} maxColumns={3} minColumnWidth={220}>{items.map((ticket) => <button key={ticket.id} type="button" className={styles.feedButton} aria-haspopup="dialog" onClick={() => setSelectedId(ticket.id)}>
					<div className={styles.inline}><strong>{ticket.subject}</strong><Badge tone={ticket.status === 'ANSWERED' ? 'success' : 'neutral'}>{statusLabels[ticket.status]}</Badge></div>
					<span className={styles.muted}>{formatDateTime(ticket.updatedAt)} · {ticket.id}</span>
				</button>)}</WidgetPreview>}
			<Link className={styles.link} to={ROUTES.support}>Открыть поддержку ↗</Link>
		</WidgetState>
		{selected && <Modal title={selected.subject} onClose={() => setSelectedId(null)} footer={<Link className={styles.link} to={paths.support(selected.applicationId ?? undefined)}>Открыть поддержку ↗</Link>}>
			<div className={styles.events}><Badge tone={selected.status === 'ANSWERED' ? 'success' : 'neutral'}>{statusLabels[selected.status]}</Badge>
				<p className={styles.muted}>{selected.id} · {formatDateTime(selected.updatedAt)}</p>
				{selected.text && <div><h3>Ваше обращение</h3><p className={styles.longText}>{selected.text}</p></div>}
				<div><h3>Ответ банка</h3><p className={styles.longText}>{selected.reply ?? 'Ответа пока нет.'}</p></div>
			</div>
		</Modal>}
	</>;
};
