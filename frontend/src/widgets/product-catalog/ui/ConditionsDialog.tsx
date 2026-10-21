import { buildConditionRows, useGetGoalsQuery, useGetProductQuery } from '@entities/product';
import { PaymentEstimate } from '@features/calculate-payment';
import { StartApplicationButton } from '@features/start-application';
import { formatMoney, formatTerm } from '@shared/lib/format';
import { AsyncBoundary, KeyValueList, Modal } from '@shared/ui';

import { estimateParams } from '../model/estimate';
import styles from './ConditionsDialog.module.scss';

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
	const goals = useGetGoalsQuery();
	const details = product.data;
	const hasTerm = details?.term !== null && details?.term !== undefined;

	return (
		<Modal title={details?.name ?? 'Условия продукта'} onClose={onClose} wide>
			<AsyncBoundary
				isLoading={product.isLoading}
				isError={product.isError}
				onRetry={() => void product.refetch()}
				loadingLabel="Загрузка условий"
				errorText="Не удалось загрузить условия."
			>
				{details && (
					<div className={styles.body}>
						<p className={styles.text}>{details.description}</p>
						<div className={styles.calc}>
							{hasTerm ? (
								<>
									<p className={styles.calcTitle}>
										Ваш расчёт: {formatMoney(amount)} на {term ? formatTerm(term) : '—'}
									</p>
									<PaymentEstimate params={estimateParams(details.code, amount, term)} />
								</>
							) : (
								<>
									<p className={styles.calcTitle}>Лимит на 12 месяцев</p>
									<p className={styles.text}>
										Проценты начисляются только за дни, когда счёт в минусе.
									</p>
								</>
							)}
						</div>
						<KeyValueList items={buildConditionRows(details, goals.data ?? [])} />
						<h3 className={styles.subTitle}>Что понадобится</h3>
						<ul className={styles.need}>
							{details.whatYouNeed.map((item) => <li key={item}>{item}</li>)}
						</ul>
						<div>
							<StartApplicationButton
								productCode={details.code}
								amount={amount}
								term={hasTerm ? term : null}
								goal={goal ?? undefined}
							/>
						</div>
					</div>
				)}
			</AsyncBoundary>
		</Modal>
	);
};
