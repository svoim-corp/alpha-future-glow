import { formatCompactMoney, formatCompactMoneyRange, formatRateRange } from '@shared/lib/format';
import { Badge, Icon, type IconName } from '@shared/ui';
import { useId } from 'react';

import { type ProductSummary } from '../model/types';
import styles from './ProductCard.module.scss';

interface ProductCardProps {
	product: ProductSummary;
	/** Подходит под задачу из подбора */
	fits: boolean;
	onSelect: (code: string) => void;
}

const ICONS: Record<string, IconName> = {
	WORKING: 'applications',
	INVEST: 'building',
	EXPRESS: 'arrow',
	OVERDRAFT: 'loans',
};

/** Карточка продукта по макету catalog.html: все карточки в ряду одной высоты, кнопка — внизу */
export const ProductCard = ({ product, fits, onSelect }: ProductCardProps) => {
	const titleId = useId();

	return (
		<article className={styles.card} aria-labelledby={titleId}>
			<div className={styles.top}>
				<span className={styles.icon}><Icon name={ICONS[product.code] ?? 'catalog'} size={20} /></span>
				<div className={styles.tags}>
					{fits && <Badge tone="success">Подходит под задачу</Badge>}
				</div>
			</div>
			<h3 id={titleId} className={styles.name}>{product.name}</h3>
			<p className={styles.text}>{product.shortText}</p>
			<div className={styles.amount}>
				{product.term
					? formatCompactMoneyRange(product.amount.min, product.amount.max)
					: `Лимит до ${formatCompactMoney(product.amount.max)}`}
			</div>
			<dl className={styles.specs}>
				<div>
					<dt className="visually-hidden">Срок</dt>
					<dd>{product.term ? `${product.term.min}–${product.term.max} мес.` : 'Лимит на 12 месяцев'}</dd>
				</div>
				<div>
					<dt className="visually-hidden">Ставка</dt>
					<dd>{formatRateRange(product.rate.min, product.rate.max)}</dd>
				</div>
			</dl>
			<button
				type="button"
				className={styles.cta}
				aria-label={`Условия: ${product.name}`}
				onClick={() => onSelect(product.code)}
			>
				Посмотреть условия <Icon name="arrow" size={17} />
			</button>
		</article>
	);
};
