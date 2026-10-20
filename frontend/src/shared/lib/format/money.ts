const rubFormatter = new Intl.NumberFormat('ru-RU', {
	style: 'currency',
	currency: 'RUB',
	maximumFractionDigits: 0,
});

/** Деньги в API — целые рубли. 1500000 → «1 500 000 ₽» */
export const formatMoney = (amount: number): string => rubFormatter.format(amount);

const pluralRules = new Intl.PluralRules('ru-RU');
const MONTH_FORMS: Record<Intl.LDMLPluralRule, string> = {
	zero: 'месяцев',
	one: 'месяц',
	two: 'месяца',
	few: 'месяца',
	many: 'месяцев',
	other: 'месяца',
};

/** 1 → «1 месяц», 3 → «3 месяца», 24 → «24 месяца», 12 → «12 месяцев» */
export const formatTerm = (months: number): string => `${months} ${MONTH_FORMS[pluralRules.select(months)]}`;
