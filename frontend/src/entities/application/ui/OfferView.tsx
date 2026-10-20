import { formatDate, formatMoney, formatRate, formatTerm } from '@shared/lib/format';
import { Alert } from '@shared/ui';

import { type Offer } from '../model/tracking';
import styles from './OfferView.module.scss';

const PRODUCT_NAME: Record<string, string> = {
	WORKING: 'Оборотный кредит',
	INVEST: 'Инвестиционный кредит',
	EXPRESS: 'Экспресс-кредит',
	OVERDRAFT: 'Овердрафт',
};

/** Решение банка: одобрение, иные условия (рядом запрошенное и предложенное) или отказ с причиной */
export const OfferView = ({ offer }: { offer: Offer }) => {
	if (offer.decision === 'REJECTED') {
		return (
			<div className={styles.offer}>
				<Alert tone="danger" title="Банк не может одобрить заявку">{offer.reason}</Alert>
				{offer.alternatives.length > 0 && (
					<>
						<p>Что можно сделать:</p>
						<ul>
							{offer.alternatives.map((item) => (
								<li key={item.productCode}>
									{`${PRODUCT_NAME[item.productCode] ?? item.productCode} — до ${formatMoney(item.maxAmount)}`}
								</li>
							))}
						</ul>
					</>
				)}
			</div>
		);
	}

	const { requested, offered } = offer;

	return (
		<div className={styles.offer}>
			{offer.decision === 'COUNTER_OFFER' && (
				<Alert tone="warning" title="Банк предлагает другие условия">{offer.explanation ?? undefined}</Alert>
			)}
			<div className={styles.compare}>
				{offer.decision === 'COUNTER_OFFER' && (
					<section className={styles.col} aria-label="Вы запрашивали">
						<h3>Вы запрашивали</h3>
						<p className={styles.amount}>{formatMoney(requested.amount)}</p>
						<p>{formatTerm(requested.term)}</p>
					</section>
				)}
				<section className={styles.col} aria-label="Предложение банка">
					<h3>{offer.decision === 'APPROVED' ? 'Одобрено' : 'Предложение банка'}</h3>
					<p className={styles.amount}>{formatMoney(offered.amount)}</p>
					<p>{`${formatTerm(offered.term)} · ${formatRate(offered.rate)}`}</p>
					<p>{`Платёж ${formatMoney(offered.monthlyPayment)} в месяц`}</p>
				</section>
			</div>
			<p>{`Предложение действует до ${formatDate(offer.validUntil)}`}</p>
		</div>
	);
};
