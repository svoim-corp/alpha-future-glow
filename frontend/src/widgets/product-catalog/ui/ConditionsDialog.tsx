import { buildConditionRows, useGetGoalsQuery, useGetProductQuery } from '@entities/product';
import { PaymentEstimate } from '@features/calculate-payment';
import { StartApplicationButton } from '@features/start-application';
import { formatMoney, formatTerm } from '@shared/lib/format';
import { Alert, AsyncBoundary, KeyValueList, Modal } from '@shared/ui';

import { estimateParams } from '../model/estimate';
import { clamp } from '../model/range';
import { selectionAdjustment } from '../model/selection';
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
	const selectedAmount = details ? clamp(amount, details.amount) : amount;
	const selectedTerm = details?.term ? clamp(term ?? details.term.min, details.term) : null;
	const adjustment = details ? selectionAdjustment(amount, term, details) : null;
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
							<AsyncBoundary
								isLoading={goals.isLoading}
								isError={goals.isError}
								onRetry={() => void goals.refetch()}
								loadingLabel="Загрузка целей продукта"
								errorText="Не удалось загрузить названия целей продукта."
							>
								<KeyValueList
									items={buildConditionRows(details, goals.data ?? [])}
								/>
							</AsyncBoundary>
							<h3 className={styles.subTitle}>Что понадобится</h3>
							<ul className={styles.need}>
								{details.whatYouNeed.map((item) => (
									<li key={item}>{item}</li>
								))}
							</ul>
						</div>
						<div className={styles.calc}>
							<p className={styles.calcTitle}>
								{details.code === 'OVERDRAFT' ? 'Лимит' : 'Ваш расчёт'}:{' '}
								{formatMoney(selectedAmount)} на{' '}
								{selectedTerm ? formatTerm(selectedTerm) : '—'}
							</p>
							{adjustment && (
								<Alert title="Параметры скорректированы">{adjustment}</Alert>
							)}
							<PaymentEstimate
								params={estimateParams(details.code, selectedAmount, selectedTerm)}
								showTotals
							/>
							{details.code === 'OVERDRAFT' && (
								<p className={styles.text}>
									Предварительный расчёт сервера на срок действия лимита.
									Фактические платежи зависят от использования лимита и условий
									договора.
								</p>
							)}
							{incompatibleGoal && (
								<Alert
									tone="warning"
									title="Продукт не подходит для выбранной цели."
								>
									Выберите другую цель или продукт в каталоге.
								</Alert>
							)}
							<div>
								<StartApplicationButton
									productCode={details.code}
									amount={incompatibleGoal ? null : selectedAmount}
									term={selectedTerm}
									goal={goal ?? undefined}
								/>
							</div>
							<p className={styles.text}>
								Откроется черновик заявки с указанными параметрами. Окончательные
								условия определятся после рассмотрения.
							</p>
						</div>
					</div>
				)}
			</AsyncBoundary>
		</Modal>
	);
};
