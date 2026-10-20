import { skipToken } from '@reduxjs/toolkit/query';
import { getErrorMessage } from '@shared/api';
import { formatMoneyRange, formatRateRange } from '@shared/lib/format';

import { type CalculatePaymentArgs, useCalculatePaymentQuery } from '../api/calculatorApi';
import styles from './PaymentEstimate.module.scss';

interface PaymentEstimateProps {
	/** null — параметры ещё не введены или невалидны: запрос не отправляем */
	params: CalculatePaymentArgs | null;
	/** Короткий вариант для карточки продукта */
	compact?: boolean;
}

export const PaymentEstimate = ({ params, compact = false }: PaymentEstimateProps) => {
	const { data, isFetching, isError, error } = useCalculatePaymentQuery(params ?? skipToken);

	if (!params) {
		return (
			<p className={styles.estimate}>
				<span className={styles.note}>Укажите сумму и срок, чтобы увидеть платёж.</span>
			</p>
		);
	}

	if (isFetching && !data) {
		return <div className={styles.line} role="status" aria-label="Считаем платёж" />;
	}

	if (isError || !data) {
		return (
			<p className={styles.estimate} role="alert">
				<span className={styles.bad}>{getErrorMessage(error, 'Не удалось рассчитать платёж.')}</span>
			</p>
		);
	}

	if (!data.withinProductLimits) {
		return (
			<p className={styles.estimate}>
				<span className={styles.bad}>Сумма или срок вне рамок продукта</span>
				{!compact && <span className={styles.note}>Измените сумму или срок либо выберите другой продукт.</span>}
			</p>
		);
	}

	return (
		<p className={styles.estimate}>
			<span className={styles.main}>≈ {formatMoneyRange(data.payment.min, data.payment.max)} в месяц</span>
			{data.gracePayment && (
				<span>
					Первые {data.gracePayment.months} мес. ≈{' '}
					{formatMoneyRange(data.gracePayment.min, data.gracePayment.max)} — только проценты
				</span>
			)}
			{!compact && (
				<span className={styles.note}>
					Ставка {formatRateRange(data.rate.min, data.rate.max)}: точную назовём после проверки,
					платёж уточнится.
				</span>
			)}
		</p>
	);
};
