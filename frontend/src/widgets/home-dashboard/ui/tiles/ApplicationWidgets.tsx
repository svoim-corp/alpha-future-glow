import { applicationHref, StatusBadge, useGetAllMyApplicationsQuery, useGetApplicationQuery } from '@entities/application';
import { type DashboardWidget } from '@entities/dashboard';
import { getErrorStatus } from '@shared/api';
import { ROUTES } from '@shared/config';
import { formatMoney } from '@shared/lib/format';
import { Button } from '@shared/ui';
import { Link } from 'react-router';

import styles from '../Dashboard.module.scss';
import { WidgetPreview } from './WidgetPreview';
import { WidgetState } from './WidgetState';

const freshness = { refetchOnFocus: true, refetchOnReconnect: true, refetchOnMountOrArgChange: true };
export const ApplicationsWidget = ({ widget }: { widget: Extract<DashboardWidget, { type: 'applications' }> }) => {
	const query = useGetAllMyApplicationsQuery(undefined, freshness);
	const items = (query.data ?? []).filter((item) => widget.settings.filter !== 'attention' || item.actionRequired);
	return <WidgetState loading={query.isLoading || (query.isFetching && !query.currentData)} error={query.isError} retry={() => void query.refetch()} empty={items.length === 0}
		emptyText={widget.settings.filter === 'attention' ? 'Заявок, требующих действия, нет.' : 'Заявок пока нет. Начните с каталога продуктов.'}>
		<WidgetPreview rowHeight={106} maxColumns={3} minColumnWidth={180}>{items.map((application) => <Link key={application.id} to={applicationHref(application)} className={styles.applicationCard}>
			<strong>{application.product.name}</strong><span>{formatMoney(application.amount)}</span><StatusBadge status={application.status} />
		</Link>)}</WidgetPreview>
		<div className={styles.inline}><span className={styles.muted}>Заявок: {items.length}</span><Link className={styles.link} to={ROUTES.applications}>Все заявки ↗</Link></div>
	</WidgetState>;
};

export const ApplicationWidget = ({ widget, onConfigure }: { widget: Extract<DashboardWidget, { type: 'application' }>; onConfigure: () => void }) => {
	const id = widget.settings.applicationId;
	const query = useGetApplicationQuery(id ?? '', { ...freshness, skip: !id });
	if (!id) return <div className={styles.state}><p className={styles.muted}>Выберите заявку для этого виджета.</p><Button variant="secondary" compact onClick={onConfigure}>Выбрать заявку</Button></div>;
	if (getErrorStatus(query.error) === 404) return <div className={styles.state}><p className={styles.muted}>Выбранная заявка больше недоступна.</p><Button variant="secondary" compact onClick={onConfigure}>Выбрать другую</Button></div>;
	const application = query.currentData;
	return <WidgetState loading={query.isLoading || (query.isFetching && !query.currentData)} error={query.isError} retry={() => void query.refetch()}>
		{application && <>
			<p className={styles.muted}>{application.product.name} · {application.number ?? 'Черновик'}</p>
			<div className={styles.number}>{formatMoney(application.params.amount)}</div><StatusBadge status={application.status} />
			<Link className={styles.link} to={applicationHref(application)}>{application.status === 'DRAFT' ? 'Продолжить заявку' : 'Открыть заявку'} ↗</Link>
		</>}
	</WidgetState>;
};

/** Вне настраиваемой сетки: скрытие виджета не прячет запрос банка. */
export const DashboardAttention = () => {
	const query = useGetAllMyApplicationsQuery(undefined, freshness);
	const applications = query.data?.filter((item) => item.actionRequired) ?? [];
	if (query.isError) return <div className={styles.attention} role="status"><span>Не удалось проверить действия по заявкам.</span><Button variant="ghost" compact onClick={() => void query.refetch()}>Повторить</Button></div>;
	const first = applications[0];
	return first ? <aside className={styles.attention} aria-label="Требует внимания"><strong>Нужно ваше действие</strong><span>Заявок, требующих внимания: {applications.length}</span><Link to={applicationHref(first)}>Открыть заявку ↗</Link></aside> : null;
};
