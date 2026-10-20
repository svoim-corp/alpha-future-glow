import { ROUTES } from '@shared/config';
import { formatMoney } from '@shared/lib/format';
import { ButtonLink, Card } from '@shared/ui';

import { useGetPlanRecommendationsQuery } from '../api/planApi';

/** Виджет главной «Идеи становятся ближе»: показывается только если по плану есть подходящий продукт */
export const PlanRecommendations = () => {
	const { data } = useGetPlanRecommendationsQuery();

	if (!data || data.length === 0) {
		return null;
	}

	return (
		<Card aria-labelledby="plan-recs">
			<h2 id="plan-recs">Идеи становятся ближе</h2>
			<ul>
				{data.map((item) => (
					<li key={item.planId}>
						<strong>{item.planTitle}</strong>{` · ${formatMoney(item.budgetRub)}. `}
						{`${item.productName} — вариант для изучения: ${item.reason}.`}
					</li>
				))}
			</ul>
			<p>Это предварительное соответствие по параметрам, а не решение банка.</p>
			<ButtonLink variant="secondary" to={ROUTES.planner}>Открыть планер</ButtonLink>
		</Card>
	);
};
