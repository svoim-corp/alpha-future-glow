import { type Application, isVersionConflict, SavedIndicator, useDraftPatch } from '@entities/application';
import { zodResolver } from '@hookform/resolvers/zod';
import { getErrorMessage } from '@shared/api';
import { Alert, Field, TextAreaField } from '@shared/ui';
import { type ReactNode, useState } from 'react';
import { useForm } from 'react-hook-form';

import { buildParamsPatch, toFormValues } from '../model/diff';
import { paramsSchema,type ParamsValues } from '../model/schema';
import styles from './EditDraftParams.module.scss';

interface EditDraftParamsProps {
	application: Application;
	/** Расчёт платежа под формой: собирается выше, потому что фичи не импортируют друг друга */
	estimate?: ReactNode;
	onChange?: (values: ParamsValues) => void;
}

/** Сумма, срок и цель. Сохраняется на уходе из поля, пока значения валидны и отличаются от сохранённых */
export const EditDraftParams = ({ application, estimate, onChange }: EditDraftParamsProps) => {
	const { save, isSaving } = useDraftPatch(application.id, application.version);
	const [failure, setFailure] = useState<string | null>(null);
	const [conflict, setConflict] = useState(false);
	const [invalidated, setInvalidated] = useState(false);
	const locked = application.locks.paramsLocked;
	const code = application.product.code;

	const { register, handleSubmit, getValues, formState: { errors } } = useForm<ParamsValues>({
		resolver: zodResolver(paramsSchema),
		defaultValues: toFormValues(application.params),
	});

	const persist = handleSubmit(async (values) => {
		const body = buildParamsPatch(values, application.params);

		if (!body) {
			return;
		}

		setFailure(null);

		try {
			const result = await save({ params: body });
			setInvalidated(result.prelimInvalidated);
		} catch (error) {
			if (isVersionConflict(error)) {
				setConflict(true);
			} else {
				setFailure(getErrorMessage(error, 'Не удалось сохранить черновик'));
			}
		}
	});

	const fire = () => {
		onChange?.(getValues());
		void persist();
	};

	return (
		<form className={styles.form} onSubmit={(event) => event.preventDefault()} onBlur={fire} noValidate>
			{conflict && (
				<Alert tone="warning" title="Заявка изменена в другом окне">
					Обновите страницу, чтобы увидеть актуальные данные, и повторите правку.
				</Alert>
			)}
			{locked && <Alert tone="info" title="Параметры зафиксированы после предварительного решения" />}
			<div className={styles.grid}>
				<Field
					label="Сумма, ₽"
					inputMode="numeric"
					disabled={locked}
					error={errors.amount?.message}
					{...register('amount')}
				/>
				<Field
					label="Срок, мес."
					inputMode="numeric"
					disabled={locked}
					error={errors.term?.message}
					{...register('term')}
				/>
				{code === 'INVEST' && (
					<Field
						label="Отсрочка основного долга, мес."
						inputMode="numeric"
						disabled={locked}
						error={errors.graceMonths?.message}
						{...register('graceMonths')}
					/>
				)}
				{code === 'OVERDRAFT' && (
					<Field
						label="Лимит овердрафта, ₽"
						inputMode="numeric"
						disabled={locked}
						error={errors.overdraftLimit?.message}
						{...register('overdraftLimit')}
					/>
				)}
			</div>
			<TextAreaField
				label="На что нужны деньги"
				disabled={locked}
				error={errors.purposeText?.message}
				{...register('purposeText')}
			/>
			{estimate}
			{invalidated && (
				<Alert tone="warning" title="Параметры изменились">
					Предварительное решение нужно получить заново.
				</Alert>
			)}
			<SavedIndicator savedAt={application.savedAt} saving={isSaving} failed={failure !== null} />
			{failure && <Alert tone="danger" title={failure} />}
		</form>
	);
};
