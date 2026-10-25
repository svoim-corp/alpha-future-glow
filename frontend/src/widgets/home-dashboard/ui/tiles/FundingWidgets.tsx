import { type DashboardWidget } from '@entities/dashboard';
import { formatPeriod, useGetPlansQuery } from '@entities/plan';
import { useGetProductsQuery } from '@entities/product';
import { ROUTES } from '@shared/config';
import { formatMoney, formatRateRange } from '@shared/lib/format';
import { Button } from '@shared/ui';
import { Link } from 'react-router';

import styles from '../Dashboard.module.scss';
import { WidgetState } from './WidgetState';

const freshness = { refetchOnFocus: true, refetchOnReconnect: true, refetchOnMountOrArgChange: true };
export const PlanFundingWidget = ({ widget, onConfigure }: { widget: Extract<DashboardWidget, { type: 'planFunding' }>; onConfigure: () => void }) => {
	const id = widget.settings.planId;
	const query = useGetPlansQuery(undefined, { ...freshness, skip: id === null });
	if (!id) return <div className={styles.state}><p className={styles.muted}>Выберите план, для которого хотите видеть вариант финансирования.</p><Button variant="secondary" compact onClick={onConfigure}>Выбрать план</Button></div>;
	const plan = query.currentData?.items.find((item) => item.id === id);
	return <WidgetState loading={query.isLoading} error={query.isError} retry={() => void query.refetch()}>
		{(query.currentData?.items.length ?? 0) >= 200 && <p className={styles.muted} title="Получено до 200 планов. Список может быть неполным.">Выборка до 200 планов</p>}
		{plan ? <>
			<strong className={styles.previewTitle}>{plan.title}</strong><span className={styles.smallNumber}>{formatMoney(plan.budgetRub)}</span><p className={styles.muted}>Ориентир бюджета · {formatPeriod(plan)}</p>
			{plan.recommendation ? <div className={styles.fundingMatch}><strong>{plan.recommendation.productName}</strong><p className={styles.fundingReason}>{plan.recommendation.reason}</p><p className={styles.muted}>Это соответствие параметрам плана, а не одобрение банка.</p></div>
				: <p className={styles.empty}>Для этого плана пока нет рекомендации по финансированию.</p>}
			<div className={styles.inline}><Link className={styles.link} to={ROUTES.planner}>Планер ↗</Link><Link className={styles.link} to={ROUTES.products}>Каталог ↗</Link></div>
		</> : <div className={styles.state}><p>План не найден в полученном списке.</p><Button variant="secondary" compact onClick={onConfigure}>Выбрать другой</Button></div>}
	</WidgetState>;
};

export const FundingWidget = ({ widget }: { widget: Extract<DashboardWidget, { type: 'funding' }> }) => {
	const query = useGetProductsQuery(undefined, freshness);
	return <WidgetState loading={query.isLoading} error={query.isError} retry={() => void query.refetch()}>
		<strong className={styles.fundingTitle}>Финансирование бизнеса</strong>
		{query.currentData?.length ? <div className={styles.fundingProducts}>{query.currentData.slice(0, widget.w >= 6 ? 3 : 2).map((product) => <p key={product.code}><strong>{product.name}</strong><span>{formatRateRange(product.rate.min, product.rate.max)}</span></p>)}</div>
			: <p className={styles.muted}>Продуктов в каталоге пока нет.</p>}
		<Link className={styles.link} to={ROUTES.products}>Подобрать финансирование ↗</Link><p className={styles.muted}>Условия и доступность зависят от выбранного продукта и решения банка.</p>
	</WidgetState>;
};
