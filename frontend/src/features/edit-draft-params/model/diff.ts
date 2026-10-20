import { type ApplicationParams, type ApplicationPatch } from '@entities/application';

import { type ParamsValues } from './schema';

export const toFormValues = (params: ApplicationParams): ParamsValues => ({
	amount: String(params.amount),
	term: params.term === null ? '' : String(params.term),
	graceMonths: params.graceMonths === null ? '' : String(params.graceMonths),
	overdraftLimit: params.overdraftLimit === null ? '' : String(params.overdraftLimit),
	purposeText: params.purposeText ?? '',
});

const toNumberOrNull = (value: string): number | null => (value === '' ? null : Number(value));

/** В PATCH уходят только поля, которые отличаются от последнего сохранённого состояния */
export const buildParamsPatch = (values: ParamsValues, saved: ApplicationParams): ApplicationPatch['params'] => {
	const next: ApplicationParams = {
		...saved,
		amount: Number(values.amount),
		term: toNumberOrNull(values.term),
		graceMonths: toNumberOrNull(values.graceMonths),
		overdraftLimit: toNumberOrNull(values.overdraftLimit),
		purposeText: values.purposeText === '' ? null : values.purposeText,
	};
	const keys = ['amount', 'term', 'graceMonths', 'overdraftLimit', 'purposeText'] as const;
	const changed: NonNullable<ApplicationPatch['params']> = {};

	for (const key of keys) {
		if (next[key] !== saved[key]) {
			Object.assign(changed, { [key]: next[key] });
		}
	}

	return Object.keys(changed).length > 0 ? changed : undefined;
};
