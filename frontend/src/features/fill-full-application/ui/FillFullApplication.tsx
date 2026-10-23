import {
	type Application,
	type GuarantorCandidate,
	isVersionConflict,
	SavedIndicator,
	useDraftPatch,
	useGetApplicationQuery,
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
	const { isFetching: refreshing } = useGetApplicationQuery(application.id);
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
	const busy = isSaving || refreshing || consentState.isLoading || validateState.isLoading || submitState.isLoading;
	const guarantorReady = personId !== '' && application.stage2.guarantor?.personId === personId;
	const guarantorConsented = guarantorReady && application.stage2.guarantor?.bkiConsent === true;

	const pickGuarantor = async (value: string) => {
		if (value === '' || busy || conflict) {
			return;
		}

		setPersonId(value);
		setSaveFailed(false);
		setMissing([]);
		setFailure(null);

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
		if (!guarantorReady || guarantorConsented || busy || saveFailed || conflict) return;
		setFailure(null);
		setMissing([]);

		try {
			await consent(application.id).unwrap();
		} catch (error) {
			setFailure(getErrorMessage(error, 'Не удалось сохранить согласие поручителя'));
		}
	};

	const send = async () => {
		if (busy || saveFailed || conflict || (personId !== '' && !guarantorReady)) return;
		setFailure(null);
		setMissing([]);

		try {
			const check = await validate(application.id).unwrap();
			// Подтверждение достоверности передаётся только в submit, поэтому validate всегда считает его недостающим.
			const rest = check.missing.filter((item) => item.field !== 'stage2.truthConfirmed');

			if (rest.length > 0 || !truth) {
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
						disabled={busy || conflict}
						error={saveFailed ? 'Не удалось сохранить поручителя. Выберите его повторно.' : undefined}
						onChange={(next) => void pickGuarantor(next)}
						options={[
							{ value: '', label: 'Выберите поручителя' },
							...candidates.map((candidate) => ({
								value: candidate.personId,
								label: candidateLabel(candidate),
							})),
						]}
					/>
					{guarantorConsented ? (
						<Alert tone="success" title="Поручитель согласился на запрос его кредитной истории" />
					) : (
						<CheckboxField
							label="Поручитель согласен на запрос его кредитной истории"
							checked={false}
							disabled={!guarantorReady || busy || saveFailed || conflict}
							onChange={() => void giveGuarantorConsent()}
						/>
					)}
				</div>
			)}
			<CheckboxField
				label="Подтверждаю, что все сведения достоверны"
				checked={truth}
				disabled={busy || conflict}
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
				<Button onClick={() => void send()} loading={validateState.isLoading || submitState.isLoading}
					disabled={isSaving || refreshing || consentState.isLoading || saveFailed || conflict || (personId !== '' && !guarantorReady)}>
					Отправить заявку
				</Button>
			</div>
		</form>
	);
};
