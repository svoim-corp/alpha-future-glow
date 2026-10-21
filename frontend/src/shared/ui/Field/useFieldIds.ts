import { useId } from 'react';

interface FieldIdsInput {
	id?: string;
	hint?: string;
	error?: string;
	describedBy?: string;
}

/** id поля и aria-describedby, связывающий подсказку и ошибку с контролом */
export const useFieldIds = ({ id, hint, error, describedBy }: FieldIdsInput) => {
	const generated = useId();
	const fieldId = id ?? generated;
	const hintId = hint ? `${fieldId}-hint` : undefined;
	const errorId = error ? `${fieldId}-error` : undefined;

	return {
		fieldId,
		hintId,
		errorId,
		describedBy: [describedBy, hintId, errorId].filter(Boolean).join(' ') || undefined,
	};
};
