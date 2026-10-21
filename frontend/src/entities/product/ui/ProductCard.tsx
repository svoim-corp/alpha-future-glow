import { formatMoney, formatRateRange } from '@shared/lib/format';
import { Badge, Button } from '@shared/ui';
import clsx from 'clsx';
import { type ReactNode, useId } from 'react';

import { type ProductSummary } from '../model/types';
import styles from './ProductCard.module.scss';

interface ProductCardProps {
	product: ProductSummary;
	selected: boolean;
	/** Подходит под задачу из «Подбора» */
	fits: boolean;
	onSelect: (code: string) => void;
	/** Расчёт платежа под сумму и срок из подбора */
	estimate?: ReactNode;
}

export const ProductCard = ({ product, selected, fits, onSelect, estimate }: ProductCardProps) => {
	const titleId = useId();

	return (
		<article
			className={clsx(styles.card, fits && styles.fits, selected && styles.selected)}
			aria-labelledby={titleId}
		>
			<div className={styles.head}>
				<h3 id={titleId} className={styles.name}>{product.name}</h3>
				{fits && <Badge tone="success">Подходит под задачу</Badge>}
			</div>
			<p className={styles.tagline}>{product.shortText}</p>
			<dl className={styles.keys}>
				<div className={styles.key}>
					<dt>{product.term ? 'Сумма' : 'Лимит'}</dt>
					<dd>до {formatMoney(product.amount.max)}</dd>
				</div>
				{product.term && (
					<div className={styles.key}>
						<dt>Срок</dt>
						<dd>{product.term.min}–{product.term.max} мес.</dd>
					</div>
				)}
				<div className={styles.key}>
					<dt>Ставка</dt>
					<dd>{formatRateRange(product.rate.min, product.rate.max)}</dd>
				</div>
			</dl>
			<div className={styles.tags}>
				{product.noStatementThreshold !== null && (
					<Badge>Без выписки до {formatMoney(product.noStatementThreshold)}</Badge>
				)}
				{product.bankClientsOnly && <Badge tone="info">Только клиентам банка</Badge>}
			</div>
			{estimate}
			<Button
				variant={selected ? 'primary' : 'secondary'}
				compact
				aria-pressed={selected}
				aria-label={`Условия: ${product.name}`}
				onClick={() => onSelect(product.code)}
			>
				Условия
			</Button>
		</article>
	);
};
