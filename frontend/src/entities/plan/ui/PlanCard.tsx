import { formatDate, formatMoney } from '@shared/lib/format';
import { Badge, Icon, type IconName } from '@shared/ui';
import { type ReactNode } from 'react';

import { type BusinessPlan } from '../model/types';
import styles from './PlanCard.module.scss';

export const formatPeriod = (plan: BusinessPlan): string => {
	if (plan.dateMode === 'UNSCHEDULED' || !plan.startDate) {
		return 'Без даты · идея для развития';
	}

	if (plan.dateMode === 'DAY' || plan.startDate === plan.endDate) {
		return formatDate(plan.startDate);
	}

	return `${formatDate(plan.startDate)} — ${plan.endDate ? formatDate(plan.endDate) : '…'}`;
};

/** Иконка плана по цели: недвижимость и развитие — здание, оборотные — заявка, оборудование — каталог */
export const planIcon = (plan: BusinessPlan): IconName => {
	if (plan.goal === 'REAL_ESTATE' || plan.goal === 'EXPANSION') {
		return 'building';
	}

	if (plan.goal === 'WORKING_CAPITAL') {
		return 'applications';
	}

	return plan.goal === 'EQUIPMENT' ? 'catalog' : 'calendar';
};

interface PlanCardProps {
	plan: BusinessPlan;
	goalName?: string;
	actions?: ReactNode;
	onOpen?: () => void;
}

/** Карточка плана в списках «Все планы» и «Идеи без даты» (макет planner.html) */
export const PlanCard = ({ plan, goalName, actions, onOpen }: PlanCardProps) => (
	<article className={styles.card} aria-label={plan.title}>
		<div className={styles.head}>
			<span className={styles.icon}><Icon name={planIcon(plan)} size={20} /></span>
			<Badge tone={plan.dateMode === 'UNSCHEDULED' ? 'neutral' : 'info'}>
				{plan.dateMode === 'UNSCHEDULED' ? 'Идея без даты' : 'План с периодом'}
			</Badge>
		</div>
		<h3>{plan.title}</h3>
		<p className={styles.period}>{`${formatPeriod(plan)} · ${formatMoney(plan.budgetRub)}`}</p>
		{goalName && <p className={styles.goal}>{`Цель: ${goalName}`}</p>}
		{plan.note && <p className={styles.note}>{plan.note}</p>}
		{plan.recommendation && (
			<p className={styles.match}>{`Подходит: ${plan.recommendation.productName}. ${plan.recommendation.reason}.`}</p>
		)}
		{actions && <div className={styles.actions}>{actions}</div>}
		{onOpen && <button type="button" className={styles.open} aria-label={`Открыть план: ${plan.title}`} aria-haspopup="dialog" onClick={onOpen} />}
	</article>
);
