import { type Application, isVersionConflict, SavedIndicator, useDraftPatch } from '@entities/application';
import { zodResolver } from '@hookform/resolvers/zod';
import { getApiError, getErrorMessage, getFieldErrors } from '@shared/api';
import { Alert, Button, CheckboxField, Field, TextAreaField } from '@shared/ui';
import { useState } from 'react';
import { useForm, useWatch } from 'react-hook-form';

import {
	type MissingField,
	useConsentGuarantorBkiMutation,
	useSubmitApplicationMutation,
	useValidateApplicationMutation,
} from '../api/submitApi';
import { buildStage2Patch, toStage2Values } from '../model/diff';
import { stage2Schema,type Stage2Values } from '../model/schema';
import styles from './FillFullApplication.module.scss';

interface FillFullApplicationProps {
	application: Application;
	onSubmitted: (redirect: string) => void;
}

/** Этап 2: фактический адрес, поручитель (для ООО), подтверждение достоверности и отправка */
export const FillFullApplication = ({ application, onSubmitted }: FillFullApplicationProps) => {
	const { save, isSaving } = useDraftPatch(application.id, application.version);
	const [consent, consentState] = useConsentGuarantorBkiMutation();
	const [validate, validateState] = useValidateApplicationMutation();
	const [submit, submitState] = useSubmitApplicationMutation();
	const [truth, setTruth] = useState(application.stage2.truthConfirmed);
	const [missing, setMissing] = useState<MissingField[]>([]);
	const [failure, setFailure] = useState<string | null>(null);
	const [saveFailed, setSaveFailed] = useState(false);
	const [conflict, setConflict] = useState(false);
	const [idempotencyKey] = useState(() => crypto.randomUUID());
	const isLlc = application.company?.type === 'LLC';

	const { register, handleSubmit, control, formState: { errors } } = useForm<Stage2Values>({
		resolver: zodResolver(stage2Schema),
		defaultValues: toStage2Values(application.stage2),
	});
	const same = useWatch({ control, name: 'sameAsRegistration' });

	const persist = handleSubmit(async (values) => {
		const body = buildStage2Patch(values, application.stage2);

		if (!body) {
			return;
		}

		setSaveFailed(false);

		try {
			await save({ stage2: body });
		} catch (error) {
			if (isVersionConflict(error)) {
				setConflict(true);
			} else {
				setSaveFailed(true);
			}
		}
	});

	const giveGuarantorConsent = async () => {
		setFailure(null);

		try {
			await consent(application.id).unwrap();
		} catch (error) {
			setFailure(getErrorMessage(error, 'Не удалось сохранить согласие поручителя'));
		}
	};

	const send = async () => {
		setFailure(null);
		setMissing([]);

		try {
			await persist();
			const check = await validate(application.id).unwrap();

			if (!check.valid || !truth) {
				const rest = check.missing.filter((item) => item.field !== 'stage2.truthConfirmed' || !truth);
				setMissing(truth ? rest : [
					...rest.filter((item) => item.field !== 'stage2.truthConfirmed'),
					{ field: 'stage2.truthConfirmed', message: 'Подтвердите достоверность сведений' },
				]);

				return;
			}

			const done = await submit({
				applicationId: application.id,
				version: application.version,
				idempotencyKey,
			}).unwrap();
			onSubmitted(done.redirect);
		} catch (error) {
			const api = getApiError(error);

			if (api?.code === 'VALIDATION_ERROR') {
				setMissing(Object.entries(getFieldErrors(error)).map(([field, message]) => ({ field, message })));
			} else {
				setFailure(getErrorMessage(error, 'Не удалось отправить заявку. Повторите попытку.'));
			}
		}
	};

	const guarantorConsented = application.stage2.guarantor?.bkiConsent === true;
	const guarantorReady = application.stage2.guarantor !== null;

	return (
		<form
			className={styles.form}
			onSubmit={(event) => event.preventDefault()}
			onBlur={() => void persist()}
			noValidate
		>
			{conflict && (
				<Alert tone="warning" title="Заявка изменена в другом окне">
					Обновите страницу, чтобы увидеть актуальные данные.
				</Alert>
			)}
			<div className={styles.group}>
				<h3 className={styles.legend}>Фактический адрес</h3>
				<CheckboxField label="Совпадает с адресом регистрации" {...register('sameAsRegistration')} />
				{!same && (
					<TextAreaField
						label="Фактический адрес"
						error={errors.address?.message}
						{...register('address')}
					/>
				)}
			</div>
			{isLlc && (
				<div className={styles.group}>
					<h3 className={styles.legend}>Поручитель</h3>
					<Field label="ФИО поручителя" error={errors.guarantorName?.message} {...register('guarantorName')} />
					<Field
						label="ИНН поручителя"
						inputMode="numeric"
						error={errors.guarantorInn?.message}
						{...register('guarantorInn')}
					/>
					<CheckboxField
						label="Поручитель согласен на запрос его кредитной истории"
						checked={guarantorConsented}
						disabled={!guarantorReady || guarantorConsented || consentState.isLoading}
						onChange={() => void giveGuarantorConsent()}
					/>
				</div>
			)}
			<CheckboxField
				label="Подтверждаю, что все сведения достоверны"
				checked={truth}
				onChange={(event) => setTruth(event.target.checked)}
			/>
			{missing.length > 0 && (
				<Alert tone="warning" title="Заполните недостающее">
					<ul className={styles.list}>
						{missing.map((item) => <li key={item.field}>{item.message}</li>)}
					</ul>
				</Alert>
			)}
			{failure && <Alert tone="danger" title={failure} />}
			<SavedIndicator savedAt={application.savedAt} saving={isSaving} failed={saveFailed} />
			<div className={styles.actions}>
				<Button onClick={() => void send()} loading={validateState.isLoading || submitState.isLoading}>
					Отправить заявку
				</Button>
			</div>
		</form>
	);
};
