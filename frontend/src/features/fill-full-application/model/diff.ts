import { type ApplicationPatch, type ApplicationStage2 } from '@entities/application';

import { type Stage2Values } from './schema';

export const toStage2Values = (stage2: ApplicationStage2): Stage2Values => ({
	sameAsRegistration: stage2.actualAddress?.sameAsRegistration ?? true,
	address: stage2.actualAddress?.address ?? '',
	guarantorName: stage2.guarantor?.fullName ?? '',
	guarantorInn: stage2.guarantor?.inn ?? '',
});

/** В PATCH уходит только то, что отличается от сохранённого; поручитель — когда заполнены оба поля */
export const buildStage2Patch = (values: Stage2Values, saved: ApplicationStage2): ApplicationPatch['stage2'] => {
	const patch: NonNullable<ApplicationPatch['stage2']> = {};
	const before = toStage2Values(saved);

	const addressChanged = values.sameAsRegistration !== before.sameAsRegistration
		|| (!values.sameAsRegistration && values.address !== before.address);

	if (addressChanged) {
		patch.actualAddress = values.sameAsRegistration
			? { sameAsRegistration: true }
			: { sameAsRegistration: false, address: values.address };
	}

	const guarantorChanged = values.guarantorName !== before.guarantorName
		|| values.guarantorInn !== before.guarantorInn;

	if (guarantorChanged && values.guarantorName !== '' && values.guarantorInn !== '') {
		patch.guarantor = { fullName: values.guarantorName, inn: values.guarantorInn };
	}

	return Object.keys(patch).length > 0 ? patch : undefined;
};
