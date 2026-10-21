import { applicationHref, type ApplicationShort, useGetOfferQuery } from '@entities/application';
import { formatDate, formatMoney, formatRate, formatTerm } from '@shared/lib/format';
import { Badge, Icon } from '@shared/ui';
import { Link } from 'react-router';

import styles from './OfferCard.module.scss';

/** Одобренная, но ещё не оформленная заявка: условия банка и срок действия предложения */
export const OfferCard = ({ application }: { application: ApplicationShort }) => {
	const { data } = useGetOfferQuery(application.id);
	const offer = data && data.decision !== 'REJECTED' ? data.offered : null;
	const validUntil = data && data.decision !== 'REJECTED' ? data.validUntil : null;
	const title = application.status === 'COUNTER_OFFER' ? 'Банк предложил новые условия' : `${application.product.name} одобрен`;

	return (
		<section className={styles.card} aria-label={title}>
			<div className={styles.illustration}><Icon name="check" size={40} /></div>
			<div className={styles.body}>
				<span className={styles.eyebrow}>ПРЕДЛОЖЕНИЕ ПО ЗАЯВКЕ {application.number}</span>
				<h2>{title}</h2>
				<p>Это предложение банка. Кредит ещё не оформлен и деньги не выданы.</p>
				<dl className={styles.terms}>
					<div><dt>Сумма</dt><dd>{formatMoney(offer?.amount ?? application.amount)}</dd></div>
					<div><dt>Срок</dt><dd>{formatTerm(offer?.term ?? application.term)}</dd></div>
					{offer && <div><dt>Ставка</dt><dd>{formatRate(offer.rate)}</dd></div>}
				</dl>
				{validUntil && <Badge tone="warning">Предложение до {formatDate(validUntil)}</Badge>}
				<div className={styles.actions}>
					<Link to={applicationHref(application)} className={styles.button}>
						Посмотреть условия <Icon name="arrow" size={17} />
					</Link>
				</div>
			</div>
		</section>
	);
};
