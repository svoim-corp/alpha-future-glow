import { formatDate, formatMoney } from '@shared/lib/format';
import { Badge, Card } from '@shared/ui';
import { type ReactNode } from 'react';

import { type BusinessPlan } from '../model/types';

export const formatPeriod = (plan: BusinessPlan): string => {
	if (plan.dateMode === 'UNSCHEDULED' || !plan.startDate) {
		return 'Без даты · идея для развития';
	}

	if (plan.dateMode === 'DAY' || plan.startDate === plan.endDate) {
		return formatDate(plan.startDate);
	}

	return `${formatDate(plan.startDate)} — ${plan.endDate ? formatDate(plan.endDate) : '…'}`;
};

interface PlanCardProps {
	plan: BusinessPlan;
	goalName?: string;
	actions?: ReactNode;
}

export const PlanCard = ({ plan, goalName, actions }: PlanCardProps) => (
	<Card as="article" aria-label={plan.title}>
		<Badge tone={plan.dateMode === 'UNSCHEDULED' ? 'neutral' : 'info'}>
			{plan.dateMode === 'UNSCHEDULED' ? 'Идея без даты' : 'План с периодом'}
		</Badge>
		<h3>{plan.title}</h3>
		<p>{`${formatPeriod(plan)} · ${formatMoney(plan.budgetRub)}`}</p>
		{goalName && <p>{`Цель: ${goalName}`}</p>}
		{plan.note && <p>{plan.note}</p>}
		{plan.recommendation && (
			<p>{`Подходит: ${plan.recommendation.productName}. ${plan.recommendation.reason}.`}</p>
		)}
		{actions}
	</Card>
);
