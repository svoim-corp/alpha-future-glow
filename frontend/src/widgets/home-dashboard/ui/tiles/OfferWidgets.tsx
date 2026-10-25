import { applicationHref, isExpiredOffer, useGetAllMyApplicationsQuery, useGetPendingOffersQuery } from '@entities/application';
import { type DashboardWidget } from '@entities/dashboard';
import { useGetProductsQuery } from '@entities/product';
import { paths, ROUTES } from '@shared/config';
import { formatDate, formatDateTime, formatMoney, formatRate, formatTerm } from '@shared/lib/format';
import { Badge, Button } from '@shared/ui';
import { Link } from 'react-router';

import styles from '../Dashboard.module.scss';
import { WidgetPreview } from './WidgetPreview';
import { WidgetState } from './WidgetState';

const freshness = { refetchOnFocus: true, refetchOnReconnect: true, refetchOnMountOrArgChange: true };
export const DraftsWidget = ({ widget: _widget }: { widget: Extract<DashboardWidget, { type: 'drafts' }> }) => {
	const query = useGetAllMyApplicationsQuery(undefined, freshness);
	const items = (query.currentData ?? []).filter((item) => item.status === 'DRAFT');
	return <WidgetState loading={query.isLoading} error={query.isError} retry={() => void query.refetch()}>
		{items.length === 0 ? <p className={styles.empty}>Незавершённых заявок нет.</p> : <WidgetPreview rowHeight={94} maxColumns={3} minColumnWidth={210}>
			{items.map((item) => <Link key={item.id} className={styles.feedCard} to={applicationHref(item)}>
				<strong>{item.product.name}</strong><span>{formatMoney(item.amount)} · {formatTerm(item.term)}</span><small className={styles.muted}>Изменено {formatDateTime(item.updatedAt)}</small>
			</Link>)}
		</WidgetPreview>}
		<Link className={styles.link} to={items[0] ? applicationHref(items[0]) : ROUTES.products}>{items[0] ? 'Продолжить последний черновик' : 'Подобрать продукт'} ↗</Link>
	</WidgetState>;
};

export const OffersWidget = ({ widget: _widget }: { widget: Extract<DashboardWidget, { type: 'offers' }> }) => {
	const query = useGetPendingOffersQuery(undefined, freshness);
	const products = useGetProductsQuery();
	const today = new Date().toISOString().slice(0, 10);
	const entries = (query.currentData ?? []).filter((entry) => _widget.settings.filter === 'all' || !isExpiredOffer(entry, today));
	return <WidgetState loading={query.isLoading} error={query.isError} retry={() => void query.refetch()}>
		{entries.length === 0 ? <p className={styles.empty}>Подходящих предложений пока нет.</p> : <WidgetPreview rowHeight={184} maxColumns={3} minColumnWidth={220}>
			{entries.map((entry) => {
				const { application, offer } = entry;
				const signing = application.status === 'SIGNING';
				const expired = isExpiredOffer(entry, today);
				const counter = application.status === 'COUNTER_OFFER';
				const name = products.data?.find((product) => product.code === offer?.offered.productCode)?.name
					?? (offer?.offered.productCode === application.product.code ? application.product.name : 'Предложение банка');
				return <div className={styles.feedCard} key={application.id}>
					<div className={styles.inline}><strong>{name}</strong><Badge tone={expired ? 'danger' : signing ? 'info' : counter ? 'warning' : 'success'}>{expired ? 'Срок истёк' : signing ? 'Оформление' : counter ? 'Другие условия' : 'Одобрено'}</Badge></div>
					<span className={styles.muted}>Заявка {application.number ?? application.id}</span>
					{offer ? <><span className={styles.smallNumber}>{formatMoney(offer.offered.amount)}</span><span className={styles.muted}>{formatRate(offer.offered.rate)} · {formatTerm(offer.offered.term)}</span>
						{!signing && <small className={styles.muted}>Предложение до {formatDate(offer.validUntil)}</small>}</>
						: <><p className={styles.notice}>Условия банка не загрузились.</p><Button variant="secondary" compact disabled={query.isFetching} onClick={() => void query.refetch()}>Повторить загрузку условий</Button></>}
					<Link className={styles.link} to={signing ? paths.applicationSign(application.id) : applicationHref(application)}>{signing ? 'Продолжить оформление' : expired ? 'Открыть заявку' : 'Рассмотреть условия'} ↗</Link>
				</div>;
			})}
		</WidgetPreview>}
		<Link className={styles.link} to={`${ROUTES.credits}?tab=approved`}>Все предложения ↗</Link>
	</WidgetState>;
};


