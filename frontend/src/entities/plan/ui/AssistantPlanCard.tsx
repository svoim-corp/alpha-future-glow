import { formatDate, formatMoney } from '@shared/lib/format';
import { ButtonLink, Card, KeyValueList } from '@shared/ui';

import { PLAN_STATE_LABELS } from '../model/calendar';
import { type PlanState } from '../model/types';

interface AssistantPlanCardProps {
	plan: {
		id: string; title: string; goal: string; note?: string; budgetRub: number;
		startDate?: string; endDate?: string; state: PlanState;
	};
}

const GOALS: Record<string, string> = {
	EQUIPMENT: 'Купить оборудование', WORKING_CAPITAL: 'Оборотные средства', CASH_GAP: 'Кассовый разрыв',
	REAL_ESTATE: 'Помещение', EXPANSION: 'Расширение бизнеса',
};

export const AssistantPlanCard = ({ plan: p }: AssistantPlanCardProps) => (
	<Card>
		<h3>{p.title}</h3>
		<KeyValueList items={[
			{ label: 'Бюджет', value: formatMoney(p.budgetRub) },
			{ label: 'Цель', value: GOALS[p.goal] ?? p.goal },
			{ label: 'Состояние', value: PLAN_STATE_LABELS[p.state] },
			{ label: 'Начало', value: p.startDate ? formatDate(p.startDate) : 'Без даты' },
			...(p.endDate ? [{ label: 'Окончание', value: formatDate(p.endDate) }] : []),
			...(p.id ? [{ label: 'ID плана', value: p.id }] : []),
		]} />
		{p.note && <p>{p.note}</p>}
		<ButtonLink to="/planner" variant="secondary">Открыть планер</ButtonLink>
	</Card>
);
