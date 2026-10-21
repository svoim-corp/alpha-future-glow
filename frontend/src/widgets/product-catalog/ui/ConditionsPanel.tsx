import { buildConditionRows, useGetGoalsQuery, useGetProductQuery } from '@entities/product';
import { PaymentEstimate } from '@features/calculate-payment';
import { StartApplicationButton } from '@features/start-application';
import { formatMoney, formatTerm } from '@shared/lib/format';
import { AsyncBoundary, Card, KeyValueList } from '@shared/ui';

import { estimateParams } from '../model/estimate';
import styles from './ProductCatalog.module.scss';

interface ConditionsPanelProps {
	productCode: string | null;
	amount: number | null;
	term: number | null;
	goal: string | null;
}

/** Боковая панель «Условия»: условия выбранного продукта, расчёт под сумму и срок, что понадобится */
export const ConditionsPanel = ({ productCode, amount, term, goal }: ConditionsPanelProps) => {
	const product = useGetProductQuery(productCode ?? '', { skip: productCode === null });
	const goals = useGetGoalsQuery();

	if (productCode === null) {
		return (
			<Card className={styles.panel} aria-label="Условия продукта">
				<p className={styles.panelKicker}>Условия продукта</p>
				<p className={styles.panelText}>Выберите продукт в списке, чтобы увидеть условия и примерный платёж.</p>
			</Card>
		);
	}

	const details = product.data;
	const hasTerm = details?.term !== null && details?.term !== undefined;

	return (
		<Card className={styles.panel} aria-label="Условия продукта">
			<p className={styles.panelKicker}>Условия продукта</p>
			<AsyncBoundary
				isLoading={product.isLoading}
				isError={product.isError}
				onRetry={() => void product.refetch()}
				loadingLabel="Загрузка условий"
				errorText="Не удалось загрузить условия."
			>
				{details && (
					<>
						<h2 className={styles.panelTitle}>{details.name}</h2>
						<p className={styles.panelText}>{details.description}</p>
						{hasTerm ? (
							<div className={styles.calc}>
								<p className={styles.calcTitle}>
									Ваш расчёт: {amount ? formatMoney(amount) : '—'} на {term ? formatTerm(term) : '—'}
								</p>
								<PaymentEstimate params={estimateParams(details.code, amount, term)} />
							</div>
						) : (
							<div className={styles.calc}>
								<p className={styles.calcTitle}>Лимит на 12 месяцев</p>
								<p className={styles.panelText}>
									Проценты начисляются только за дни, когда счёт в минусе.
								</p>
							</div>
						)}
						<KeyValueList items={buildConditionRows(details, goals.data ?? [])} />
						<h3 className={styles.subTitle}>Что понадобится</h3>
						<ul className={styles.need}>
							{details.whatYouNeed.map((item) => <li key={item}>{item}</li>)}
						</ul>
						<StartApplicationButton
							productCode={details.code}
							amount={hasTerm ? amount : (amount ?? details.amount.min)}
							term={hasTerm ? term : null}
							goal={goal ?? undefined}
						/>
					</>
				)}
			</AsyncBoundary>
		</Card>
	);
};
