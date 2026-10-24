import { type InfoRequest } from '@entities/application';
import { getApiError, getErrorMessage } from '@shared/api';
import { Alert, Button, TextAreaField } from '@shared/ui';
import { type FormEvent, useState } from 'react';

import { useAnswerInfoRequestMutation } from '../api/answerApi';
import styles from './AnswerInfoRequest.module.scss';

interface AnswerInfoRequestProps {
	applicationId: string;
	request: InfoRequest;
}

/** Ответ на запрос банка: все пункты обязательны, заявка после ответа возвращается на проверку.
    Когда ответ уже дан (сейчас, раньше или в другой вкладке), поля только для чтения */
export const AnswerInfoRequest = ({ applicationId, request }: AnswerInfoRequestProps) => {
	const [answer, { isLoading }] = useAnswerInfoRequestMutation();
	const [values, setValues] = useState<Record<string, string>>({});
	const [errors, setErrors] = useState<Record<string, string>>({});
	const [failure, setFailure] = useState<string | null>(null);
	const [sent, setSent] = useState(false);
	const readOnly = sent || request.status === 'ANSWERED';
	const saved = Object.fromEntries((request.answers ?? []).map((item) => [item.key, item.value]));

	const submit = async (event: FormEvent) => {
		event.preventDefault();
		const next: Record<string, string> = {};

		for (const item of request.items) {
			if (!values[item.key]?.trim()) {
				next[item.key] = 'Заполните ответ';
			}
		}

		setErrors(next);

		if (Object.keys(next).length > 0) {
			return;
		}

		setFailure(null);

		try {
			await answer({
				applicationId,
				requestId: request.id,
				answers: request.items.map((item) => ({ key: item.key, value: (values[item.key] ?? '').trim() })),
			}).unwrap();
			setSent(true);
		} catch (error) {
			if (getApiError(error)?.code === 'INVALID_TRANSITION') {
				setSent(true);

				return;
			}

			setFailure(getErrorMessage(error, 'Не удалось отправить ответ. Введённое сохранено — повторите попытку.'));
		}
	};

	return (
		<form className={styles.form} onSubmit={(event) => void submit(event)} noValidate>
			<p className={styles.comment}>{request.comment}</p>
			{request.items.map((item) => (
				<TextAreaField
					key={item.key}
					label={item.label}
					value={readOnly ? (values[item.key] ?? saved[item.key] ?? '') : (values[item.key] ?? '')}
					error={errors[item.key]}
					readOnly={readOnly}
					onChange={(event) => setValues((prev) => ({ ...prev, [item.key]: event.target.value }))}
				/>
			))}
			{failure && <Alert tone="danger" title={failure} />}
			{readOnly ? (
				<Alert tone="success" title="Ответ отправлен банку" />
			) : (
				<div>
					<Button type="submit" loading={isLoading}>Отправить ответ</Button>
				</div>
			)}
		</form>
	);
};
