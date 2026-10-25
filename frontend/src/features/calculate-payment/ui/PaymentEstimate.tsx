import { skipToken } from '@reduxjs/toolkit/query';
import { getErrorMessage } from '@shared/api';
import { useDebouncedValue } from '@shared/lib/debounce';
import { formatMoneyRange, formatRateRange } from '@shared/lib/format';
import { Button } from '@shared/ui';
import { useMemo } from 'react';

import { type CalculatePaymentArgs, useCalculatePaymentQuery } from '../api/calculatorApi';
import { paymentTotals } from '../model/totals';
import styles from './PaymentEstimate.module.scss';
import { useStableHeight } from './useStableHeight';

interface PaymentEstimateProps {
	/** null — параметры ещё не введены или невалидны: запрос не отправляем */
	params: CalculatePaymentArgs | null;
	/** Короткий вариант для карточки продукта */
	compact?: boolean;
	showTotals?: boolean;
}

export const PaymentEstimate = ({
	params,
	compact = false,
	showTotals = false,
}: PaymentEstimateProps) => {
	const productCode = params?.productCode;
	const amount = params?.amount;
	const term = params?.term;
	const graceMonths = params?.graceMonths;
	const request = useMemo(
		() =>
			productCode !== undefined && amount !== undefined && term !== undefined
				? { productCode, amount, term, graceMonths }
				: null,
		[productCode, amount, term, graceMonths],
	);
	const debounced = useDebouncedValue(request);
	const {
		currentData: data,
		isFetching,
		isError,
		error,
		refetch,
	} = useCalculatePaymentQuery(debounced ?? skipToken);

	const { contentRef, minHeight } = useStableHeight();
	const pending = Boolean(params) && (request !== debounced || isFetching || (!data && !isError));

	const renderEstimate = () => {
		if (!params) {
			return (
				<p className={styles.estimate}>
					<span className={styles.note}>Укажите сумму и срок, чтобы увидеть платёж.</span>
				</p>
			);
		}

		if (request !== debounced || isFetching || (!data && !isError)) {
			return (
				<div className={styles.pending} role="status">
					<span className={styles.line} />
					Пересчитываем платёж…
				</div>
			);
		}

		if (isError || !data) {
			return (
				<div className={styles.estimate} role="alert">
					<span className={styles.bad}>
						{getErrorMessage(error, 'Не удалось рассчитать платёж.')}
					</span>
					<Button variant="secondary" compact onClick={() => void refetch()}>
						Повторить расчёт
					</Button>
				</div>
			);
		}

		if (!data.withinProductLimits) {
			return (
				<p className={styles.estimate}>
					<span className={styles.bad}>Сумма или срок вне рамок продукта</span>
					{!compact && (
						<span className={styles.note}>
							Измените сумму или срок либо выберите другой продукт.
						</span>
					)}
				</p>
			);
		}

		const totals = showTotals ? paymentTotals(params, data) : null;

		return (
			<div className={styles.estimate}>
				<span className={styles.main}>
					≈ {formatMoneyRange(data.payment.min, data.payment.max)} в месяц
				</span>
				{data.gracePayment && (
					<span>
						Первые {data.gracePayment.months} мес. ≈{' '}
						{formatMoneyRange(data.gracePayment.min, data.gracePayment.max)} — только
						проценты
					</span>
				)}
				{!compact && (
					<span className={styles.note}>
						Ставка {formatRateRange(data.rate.min, data.rate.max)}: точную назовём после
						проверки, платёж уточнится.
					</span>
				)}
				{totals && (
					<>
						<dl className={styles.totals}>
							<div>
								<dt>Общая сумма выплат</dt>
								<dd>
									≈ {formatMoneyRange(totals.payment.min, totals.payment.max)}
								</dd>
							</div>
							<div>
								<dt>Проценты за весь срок</dt>
								<dd>
									≈ {formatMoneyRange(totals.interest.min, totals.interest.max)}
								</dd>
							</div>
						</dl>
						<span className={styles.note}>
							Ориентир по округлённым платежам сервера, без комиссий. Итоговые суммы
							уточнятся в договоре.
						</span>
					</>
				)}
			</div>
		);
	};

	return (
		<div className={styles.frame} style={{ minHeight }} aria-busy={pending}>
			<div ref={contentRef}>{renderEstimate()}</div>
		</div>
	);
};
