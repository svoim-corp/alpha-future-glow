import {
	type Application,
	type GuarantorCandidate,
	isVersionConflict,
	SavedIndicator,
	useDraftPatch,
	useGetCompanyProfileQuery,
} from '@entities/application';
import { getApiError, getErrorMessage, getFieldErrors } from '@shared/api';
import { Alert, Button, CheckboxField, SelectField } from '@shared/ui';
import { useState } from 'react';

import {
	type MissingField,
	useConsentGuarantorBkiMutation,
	useSubmitApplicationMutation,
	useValidateApplicationMutation,
} from '../api/submitApi';
import styles from './FillFullApplication.module.scss';

interface FillFullApplicationProps {
	application: Application;
	onSubmitted: (redirect: string) => void;
}

const ROLE_LABELS = { HEAD: 'руководитель', FOUNDER: 'учредитель' } as const;

const candidateLabel = (candidate: GuarantorCandidate) => (
	`${candidate.fullName} — ${candidate.roles.map((role) => ROLE_LABELS[role]).join(', ')}`
);

/** Этап 2: выбор поручителя из известных банку людей (для ООО), подтверждение достоверности и отправка */
export const FillFullApplication = ({ application, onSubmitted }: FillFullApplicationProps) => {
	const { save, isSaving } = useDraftPatch(application.id, application.version);
	const { data: profile } = useGetCompanyProfileQuery(application.id);
	const [consent, consentState] = useConsentGuarantorBkiMutation();
	const [validate, validateState] = useValidateApplicationMutation();
	const [submit, submitState] = useSubmitApplicationMutation();
	const [truth, setTruth] = useState(application.stage2.truthConfirmed);
	const [personId, setPersonId] = useState(application.stage2.guarantor?.personId ?? '');
	const [missing, setMissing] = useState<MissingField[]>([]);
	const [failure, setFailure] = useState<string | null>(null);
	const [saveFailed, setSaveFailed] = useState(false);
	const [conflict, setConflict] = useState(false);
	const [idempotencyKey] = useState(() => crypto.randomUUID());
	const isLlc = application.company?.type === 'LLC';
	const candidates = profile?.guarantorCandidates ?? [];

	const pickGuarantor = async (value: string) => {
		setPersonId(value);

		if (value === '') {
			return;
		}

		setSaveFailed(false);

		try {
			await save({ stage2: { guarantor: { personId: value } } });
		} catch (error) {
			if (isVersionConflict(error)) {
				setConflict(true);
			} else {
				setSaveFailed(true);
			}
		}
	};

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
			const check = await validate(application.id).unwrap();

			if (!check.valid || !truth) {
				const rest = check.missing.filter((item) => item.field !== 'stage2.truthConfirmed');
				setMissing(truth ? rest : [
					...rest,
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
		<form className={styles.form} onSubmit={(event) => event.preventDefault()} noValidate>
			{conflict && (
				<Alert tone="warning" title="Заявка изменена в другом окне">
					Обновите страницу, чтобы увидеть актуальные данные.
				</Alert>
			)}
			{isLlc && (
				<div className={styles.group}>
					<SelectField
						label="Поручитель"
						hint="Выберите из руководителя и учредителей — их данные уже есть в банке"
						value={personId}
						onChange={(next) => void pickGuarantor(next)}
						options={[
							{ value: '', label: 'Выберите поручителя' },
							...candidates.map((candidate) => ({
								value: candidate.personId,
								label: candidateLabel(candidate),
							})),
						]}
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
