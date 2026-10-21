import { getErrorMessage } from '@shared/api';
import { Alert, Icon } from '@shared/ui';
import clsx from 'clsx';
import { useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router';

import { useGetNotificationsQuery, useMarkNotificationReadMutation } from '../api/notificationApi';
import { type Notification } from '../model/types';
import styles from './NotificationBell.module.scss';

/** Колокольчик: число непрочитанных, список; клик по уведомлению отмечает его прочитанным и ведёт по ссылке */
export const NotificationBell = () => {
	const { data, isError, error } = useGetNotificationsQuery(undefined, { pollingInterval: 60000 });
	const [markRead] = useMarkNotificationReadMutation();
	const [open, setOpen] = useState(false);
	const rootRef = useRef<HTMLDivElement>(null);
	const navigate = useNavigate();
	const unread = data?.unreadCount ?? 0;

	useEffect(() => {
		if (!open) {
			return undefined;
		}

		const onKey = (event: KeyboardEvent) => {
			if (event.key === 'Escape') {
				setOpen(false);
			}
		};
		const onClick = (event: MouseEvent) => {
			if (!rootRef.current?.contains(event.target as Node)) {
				setOpen(false);
			}
		};
		document.addEventListener('keydown', onKey);
		document.addEventListener('mousedown', onClick);

		return () => {
			document.removeEventListener('keydown', onKey);
			document.removeEventListener('mousedown', onClick);
		};
	}, [open]);

	const openItem = async (item: Notification) => {
		setOpen(false);

		if (!item.read) {
			await markRead(item.id);
		}

		void navigate(item.link);
	};

	return (
		<div className={styles.root} ref={rootRef}>
			<button
				type="button"
				className={styles.button}
				aria-expanded={open}
				aria-haspopup="true"
				aria-label={unread > 0 ? `Уведомления, непрочитанных: ${unread}` : 'Уведомления'}
				onClick={() => setOpen((prev) => !prev)}
			>
				<Icon name="bell" />
				{unread > 0 && <span className={styles.count} aria-hidden="true">{unread}</span>}
			</button>
			{open && (
				<div className={styles.panel} role="region" aria-label="Список уведомлений">
					{isError && <Alert tone="danger" title={getErrorMessage(error, 'Не удалось загрузить уведомления')} />}
					{data && data.items.length === 0 && <p>Новых уведомлений нет</p>}
					<ul className={styles.list}>
						{data?.items.map((item) => (
							<li key={item.id}>
								<button
									type="button"
									className={clsx(styles.item, !item.read && styles.unread)}
									onClick={() => void openItem(item)}
								>
									<span className={styles.title}>{item.title}</span>
									<span className={styles.text}>{item.text}</span>
								</button>
							</li>
						))}
					</ul>
				</div>
			)}
		</div>
	);
};
