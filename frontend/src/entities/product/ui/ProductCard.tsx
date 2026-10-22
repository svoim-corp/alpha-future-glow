import { formatCompactMoney, formatCompactMoneyRange, formatRateRange, formatTerm } from '@shared/lib/format';
import { AsyncBoundary, Badge, Icon, type IconName, Modal } from '@shared/ui';
import { useId, useState } from 'react';

import { useGetProductQuery } from '../api/productApi';
import { type ProductSummary } from '../model/types';
import styles from './ProductCard.module.scss';

interface ProductCardProps {
	product: ProductSummary;
	/** Подходит под задачу из подбора */
	fits: boolean;
	reason?: string;
	goalLabel?: string;
	onSelect: (code: string) => void;
}

const ICONS: Record<string, IconName> = {
	WORKING: 'applications',
	INVEST: 'building',
	EXPRESS: 'arrow',
	OVERDRAFT: 'loans',
};

/** Карточка продукта по макету catalog.html: все карточки в ряду одной высоты, кнопка — внизу */
export const ProductCard = ({ product, fits, reason, goalLabel, onSelect }: ProductCardProps) => {
	const titleId = useId();
	const [requirementsOpen, setRequirementsOpen] = useState(false);
	const details = useGetProductQuery(product.code, { skip: !requirementsOpen });
	const overdraft = product.code === 'OVERDRAFT';
	const fixedTerm = product.term && product.term.min === product.term.max;

	return (
		<article className={styles.card} aria-labelledby={titleId}>
			<div className={styles.top}>
				<span className={styles.icon}><Icon name={ICONS[product.code] ?? 'catalog'} size={20} /></span>
				<div className={styles.tags}>
					{goalLabel && <span className={styles.label}>{goalLabel}</span>}
					{fits && <Badge tone="success">Подходит под задачу</Badge>}
				</div>
			</div>
			<h3 id={titleId} className={styles.name}>{product.name}</h3>
			<p className={styles.text}>{product.shortText}</p>
			<div className={styles.amount}>
				{overdraft ? `Лимит до ${formatCompactMoney(product.amount.max)}`
					: formatCompactMoneyRange(product.amount.min, product.amount.max)}
			</div>
			<dl className={styles.specs}>
				<div>
					<dt className="visually-hidden">Срок</dt>
					<dd>{product.term
						? (fixedTerm ? `${overdraft ? 'Лимит на ' : ''}${formatTerm(product.term.min)}` : `${product.term.min}–${product.term.max} мес.`)
						: 'Срок уточняется'}</dd>
				</div>
				<div>
					<dt className="visually-hidden">Ставка</dt>
					<dd>{formatRateRange(product.rate.min, product.rate.max)}</dd>
				</div>
			</dl>
			{reason && <p className={styles.reason}>{reason}</p>}
			<button type="button" className={styles.requirementsTrigger}
				aria-haspopup="dialog" aria-label={`Что понадобится: ${product.name}`}
				onClick={() => setRequirementsOpen(true)}>
				Что понадобится <Icon name="arrow" size={16} />
			</button>
			{requirementsOpen && (
				<Modal title="Что понадобится" size="small" layer="overlay" onClose={() => setRequirementsOpen(false)}>
					<p className={styles.requirementsProduct}>{product.name}</p>
					<AsyncBoundary isLoading={details.isLoading} isError={details.isError}
						onRetry={() => void details.refetch()} loadingLabel="Загрузка требований"
						errorText="Не удалось загрузить требования."
						isEmpty={details.data?.whatYouNeed.length === 0} emptyTitle="Дополнительных действий не требуется">
						<ul className={styles.requirementsList}>
							{details.data?.whatYouNeed.map((item) => (
								<li key={item}><Icon name="check" size={18} /><span>{item}</span></li>
							))}
						</ul>
					</AsyncBoundary>
				</Modal>
			)}
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
