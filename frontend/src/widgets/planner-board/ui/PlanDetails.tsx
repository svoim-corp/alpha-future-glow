import { type BusinessPlan, formatPeriodLong, PLAN_STATE_LABELS, planPhase } from '@entities/plan';
import { PlanActions } from '@features/manage-plans';
import { formatMoney } from '@shared/lib/format';
import { Badge, Modal } from '@shared/ui';

import { PlanFunding } from './PlanFunding';
import styles from './PlannerBoard.module.scss';

interface PlanDetailsProps {
	plan: BusinessPlan;
	today: string;
	goalName?: string;
	onClose: () => void;
}

export const PlanDetails = ({ plan, today, goalName, onClose }: PlanDetailsProps) => (
	<PlanActions plan={plan} renderView={(actions) => (
		<Modal title={plan.title} onClose={onClose} size="large">
			<section className={styles.details} aria-label="Детали плана">
				<div className={styles.overview}>
					<Badge tone={plan.state === 'ACTIVE' ? 'success' : 'neutral'}>{PLAN_STATE_LABELS[plan.state]}</Badge>
					<p className={styles.period}>{formatPeriodLong(plan)}</p>
					<p className={styles.phase}>{planPhase(plan, today)}</p>
					<div className={styles.budget}>
						<small>ОРИЕНТИР БЮДЖЕТА</small>
						<strong>{formatMoney(plan.budgetRub)}</strong>
						<span>{goalName ?? plan.goal}</span>
					</div>
					{plan.note && <p className={styles.noteText}>{plan.note}</p>}
				</div>
				<section className={styles.tracking} aria-label="Финансирование плана">
					<h3>Вариант финансирования</h3>
					{plan.state !== 'ACTIVE' ? <p>В этом состоянии план не участвует в подборе.</p>
						: planPhase(plan, today) === 'Период прошёл' ? <p>Период закончился. Подбор по этому плану больше не выполняется.</p>
							: plan.recommendation ? (
								<PlanFunding key={plan.recommendation.productCode}
									productCode={plan.recommendation.productCode}
									productName={plan.recommendation.productName} />
							) : <p>Подходящего варианта пока нет. План сохранён.</p>}
				</section>
				<div className={styles.actions}>{actions}</div>
			</section>
		</Modal>
	)} />
);
