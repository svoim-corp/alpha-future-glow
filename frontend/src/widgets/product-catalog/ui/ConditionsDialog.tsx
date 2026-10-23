import { buildConditionRows, useGetGoalsQuery, useGetProductQuery } from '@entities/product';
import { Alert, AsyncBoundary, KeyValueList, Modal } from '@shared/ui';
import { useState } from 'react';

import { clamp } from '../model/range';
import { selectionAdjustment } from '../model/selection';

import styles from './ConditionsDialog.module.scss';
import { PaymentCalculator } from './PaymentCalculator';

interface ConditionsDialogProps {
	productCode: string;
	amount: number;
	term: number | null;
	goal: string | null;
	onClose: () => void;
}

/** Условия выбранного продукта: описание, расчёт под сумму и срок из калькулятора, что понадобится */
export const ConditionsDialog = ({
	productCode,
	amount,
	term,
	goal,
	onClose,
}: ConditionsDialogProps) => {
	const product = useGetProductQuery(productCode);
	const [amountValue, setAmountValue] = useState(amount);
	const [termValue, setTermValue] = useState(term);
	const goals = useGetGoalsQuery();
	const details = product.data;
	const selectedAmount = details ? clamp(amountValue, details.amount) : amountValue;
	const selectedTerm = details?.term ? clamp(termValue ?? details.term.min, details.term) : null;
	const adjustment = details ? selectionAdjustment(amountValue, termValue, details) : null;
	const incompatibleGoal = Boolean(goal && details && !details.goals.includes(goal));

	return (
		<Modal title={details?.name ?? 'Условия продукта'} onClose={onClose} size="large">
			<AsyncBoundary
				isLoading={product.isLoading}
				isError={product.isError}
				onRetry={() => void product.refetch()}
				loadingLabel="Загрузка условий"
				errorText="Не удалось загрузить условия."
			>
				{details && (
					<div className={styles.body}>
						<div className={styles.conditions}>
							<p className={styles.text}>{details.description}</p>
							<AsyncBoundary isLoading={goals.isLoading} isError={goals.isError}
								onRetry={() => void goals.refetch()} loadingLabel="Загрузка целей продукта"
								errorText="Не удалось загрузить названия целей продукта.">
								<KeyValueList items={buildConditionRows(details, goals.data ?? [])} />
							</AsyncBoundary>
							<h3 className={styles.subTitle}>Что понадобится</h3>
							<ul className={styles.need}>
								{details.whatYouNeed.map((item) => <li key={item}>{item}</li>)}
							</ul>
						</div>
						<div>
							{incompatibleGoal && <Alert tone="warning" title="Продукт не подходит для выбранной цели.">Выберите другую цель или продукт в каталоге.</Alert>}
							<PaymentCalculator compact products={[]} product={details}
								amount={selectedAmount} term={selectedTerm} goal={goal}
								adjustment={adjustment} applicationBlocked={incompatibleGoal}
								onAmountChange={setAmountValue} onTermChange={setTermValue} />
						</div>
					</div>
				)}
			</AsyncBoundary>
		</Modal>
	);
};
