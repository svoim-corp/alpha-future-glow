import { applicationHref, isExpiredOffer, type PendingOffer } from '@entities/application';
import { useGetProductQuery } from '@entities/product';
import { paths } from '@shared/config';
import { formatDate, formatMoney, formatRate, formatTerm } from '@shared/lib/format';
import { Alert, Badge, Button, ButtonLink, Icon } from '@shared/ui';

import styles from './OfferCard.module.scss';

interface OfferCardProps {
	entry: PendingOffer;
	today: string;
	onRetry: () => void;
	refreshing: boolean;
}

export const OfferCard = ({ entry, today, onRetry, refreshing }: OfferCardProps) => {
	const { application, offer } = entry;
	const product = useGetProductQuery(offer?.offered.productCode ?? '', { skip: !offer });
	const signing = application.status === 'SIGNING';
	const expired = isExpiredOffer(entry, today);
	const title = signing ? 'Оформление кредита' : offer?.decision === 'COUNTER_OFFER'
		? 'Банк предложил другие условия' : 'Кредит одобрен';

	return (
		<section className={styles.card} aria-label={title}>
			<div className={styles.illustration}><Icon name="check" size={40} /></div>
			<div className={styles.body}>
				<span className={styles.eyebrow}>ПО ЗАЯВКЕ {application.number ?? 'без номера'}</span>
				<h2>{title}</h2>
				<p>{signing ? 'Предложение принято. Продолжите оформление договора; деньги ещё не выданы.' : 'Это предложение банка. Кредит ещё не выдан.'}</p>
				{offer ? (
					<>
						{product.data && <h3>{product.data.name}</h3>}
						<dl className={styles.terms}>
							<div><dt>Сумма</dt><dd>{formatMoney(offer.offered.amount)}</dd></div>
							<div><dt>Срок</dt><dd>{formatTerm(offer.offered.term)}</dd></div>
							<div><dt>Ставка</dt><dd>{formatRate(offer.offered.rate)}</dd></div>
							<div><dt>Платёж в месяц</dt><dd>{formatMoney(offer.offered.monthlyPayment)}</dd></div>
						</dl>
						{offer.explanation && <p>{offer.explanation}</p>}
						{!signing && <Badge tone={expired ? 'danger' : 'warning'}>
							{expired ? `Срок предложения истёк ${formatDate(offer.validUntil)}` : `Предложение до ${formatDate(offer.validUntil)}`}
						</Badge>}
					</>
				) : (
					<Alert tone="danger" title={entry.error ?? 'Условия банка недоступны.'}
						actions={<Button variant="secondary" compact disabled={refreshing} onClick={onRetry}>Повторить загрузку условий</Button>}>
						Запрошенные сумма и срок не подставляются вместо условий банка.
					</Alert>
				)}
				<div className={styles.actions}>
					<ButtonLink to={signing ? paths.applicationSign(application.id) : applicationHref(application)}>
						{signing ? 'Продолжить оформление' : 'Открыть заявку'} <Icon name="arrow" size={17} />
					</ButtonLink>
				</div>
			</div>
		</section>
	);
};
