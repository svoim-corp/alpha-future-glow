import { ROUTES } from '@shared/config';
import { formatMoney } from '@shared/lib/format';
import { Icon } from '@shared/ui';
import clsx from 'clsx';
import { Link } from 'react-router';

import { useGetPlanRecommendationsQuery } from '../api/planApi';
import styles from './PlanRecommendations.module.scss';

export interface PlanRecommendationsProps {
	/** Для размещения в сетке родителя */
	className?: string;
}

/** Виджет главной «Идеи становятся ближе» (макет home.html): виден только если по плану есть подходящий продукт */
export const PlanRecommendations = ({ className }: PlanRecommendationsProps) => {
	const { data } = useGetPlanRecommendationsQuery();

	if (!data || data.length === 0) {
		return null;
	}

	return (
		<section className={clsx(styles.root, className)} aria-labelledby="plan-recs">
			<div className={styles.heading}>
				<div>
					<span className={styles.eyebrow}>ИЗ ВАШЕГО ПЛАНЕРА</span>
					<h2 id="plan-recs">Идеи становятся ближе</h2>
				</div>
				<Link to={ROUTES.planner} className={styles.textLink}>Все планы <Icon name="arrow" size={15} /></Link>
			</div>
			<ul className={styles.list}>
				{data.map((item) => (
					<li key={item.planId} className={styles.row}>
						<div className={styles.idea}>
							<span className={styles.ideaIcon}><Icon name="catalog" /></span>
							<div>
								<span className={styles.label}>ИДЕЯ ИЗ ПЛАНЕРА</span>
								<h3>{item.planTitle}</h3>
								<strong className={styles.budget}>
									{formatMoney(item.budgetRub)} <small>ориентир бюджета</small>
								</strong>
							</div>
						</div>
						<div className={styles.match}>
							<span className={styles.label}>ВАРИАНТ ДЛЯ ИЗУЧЕНИЯ</span>
							<h3>{item.productName}</h3>
							<p>{item.reason}</p>
							<div>
								<span>Подбор по цели и бюджету</span>
								<Link to={ROUTES.products}>Изучить <Icon name="arrow" size={15} /></Link>
							</div>
						</div>
					</li>
				))}
			</ul>
			<p className={styles.footnote}>Это предварительное соответствие по параметрам, а не решение банка.</p>
		</section>
	);
};
