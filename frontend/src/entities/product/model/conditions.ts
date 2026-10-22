import { formatMoney, formatRateRange, formatTerm } from '@shared/lib/format';
import { type KeyValueItem } from '@shared/ui';

import { type Goal, type ProductDetails, type Range } from './types';

const formatMoneyBounds = ({ min, max }: Range): string => `${formatMoney(min)} – ${formatMoney(max)}`;

const formatMonthsBounds = ({ min, max }: Range): string => min === max ? formatTerm(min) : `${min}–${max} мес.`;

const requirementsText = (details: ProductDetails): string => {
	const parts: string[] = [];
	const age = details.requirements.minBusinessAgeMonths;

	if (age) {
		parts.push(`бизнесу не меньше ${age} мес.`);
	}

	if (details.requirements.guarantorForLlc) {
		parts.push('для ООО нужен поручитель');
	}

	return parts.length > 0 ? parts.join('; ') : 'Особых требований нет';
};

/** Строки панели «Условия»: только то, что пришло из контракта, без домысливания */
export const buildConditionRows = (details: ProductDetails, goals: Goal[]): KeyValueItem[] => {
	const goalNames = details.goals.map((code) => goals.find((goal) => goal.code === code)?.name ?? code);
	const rows: KeyValueItem[] = [
		{ label: 'Цели', value: goalNames.join(', ') || '—' },
		{ label: details.code === 'OVERDRAFT' ? 'Размер лимита' : 'Сумма', value: formatMoneyBounds(details.amount) },
	];

	if (details.term) {
		rows.push({ label: details.code === 'OVERDRAFT' ? 'Срок действия лимита' : 'Срок', value: formatMonthsBounds(details.term) });
	}

	rows.push({ label: 'Ставка', value: formatRateRange(details.rate.min, details.rate.max) });
	rows.push({ label: 'Требования', value: requirementsText(details) });
	if (details.limitRule) {
		rows.push({ label: 'Правило определения лимита', value: details.limitRule });
	}

	if (details.graceMaxMonths) {
		rows.push({ label: 'Отсрочка', value: `По основному долгу до ${details.graceMaxMonths} мес.` });
	}

	return rows;
};
