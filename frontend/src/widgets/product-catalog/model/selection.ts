import { type ProductSummary } from '@entities/product';
import { formatMoney, formatTerm } from '@shared/lib/format';

import { clamp } from './range';

export const selectProduct = (products: ProductSummary[], code: string | null, goal: string | null) => {
	const eligible = goal ? products.filter((product) => product.goals.includes(goal)) : products;
	return eligible.find((product) => product.code === code)
		?? eligible.find((product) => product.code === 'WORKING')
		?? eligible[0];
};

export const selectionAdjustment = (amount: number, term: number | null, product: Pick<ProductSummary, 'name' | 'amount' | 'term'>): string | null => {
	const changes: string[] = [];
	const nextAmount = clamp(amount, product.amount);
	const nextTerm = product.term && term !== null ? clamp(term, product.term) : term;
	if (nextAmount !== amount) changes.push(`сумма изменена с ${formatMoney(amount)} на ${formatMoney(nextAmount)}`);
	if (term !== null && nextTerm !== null && nextTerm !== term) {
		changes.push(`срок изменён с ${formatTerm(term)} на ${formatTerm(nextTerm)}`);
	}
	return changes.length > 0
		? `Для продукта «${product.name}» ${changes.join(', ')} по его условиям. В заявку перейдут показанные параметры.`
		: null;
};
