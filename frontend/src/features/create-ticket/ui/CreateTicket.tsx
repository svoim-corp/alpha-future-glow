import { zodResolver } from '@hookform/resolvers/zod';
import { getErrorMessage } from '@shared/api';
import { Alert, Button, Field, TextAreaField } from '@shared/ui';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';

import { useCreateTicketMutation } from '../api/ticketApi';

const schema = z.object({
	subject: z.string().trim().min(3, 'Опишите тему в нескольких словах'),
	text: z.string().trim().min(10, 'Расскажите подробнее — хотя бы в одном предложении'),
});

type Values = z.infer<typeof schema>;

/** Обращение в поддержку. Со страницы заявки номер заявки подставляется сам */
export const CreateTicket = ({ applicationId }: { applicationId?: string }) => {
	const [create, { isLoading }] = useCreateTicketMutation();
	const [failure, setFailure] = useState<string | null>(null);
	const [sent, setSent] = useState(false);
	const { register, handleSubmit, reset, formState: { errors } } = useForm<Values>({
		resolver: zodResolver(schema),
		defaultValues: { subject: '', text: '' },
	});

	const onSubmit = handleSubmit(async (values) => {
		setFailure(null);
		setSent(false);

		try {
			await create({ ...values, applicationId }).unwrap();
			reset();
			setSent(true);
		} catch (error) {
			setFailure(getErrorMessage(error, 'Не удалось отправить обращение. Текст сохранён — повторите попытку.'));
		}
	});

	return (
		<form onSubmit={(event) => void onSubmit(event)} noValidate>
			{applicationId && <p>{`Обращение будет привязано к заявке ${applicationId}.`}</p>}
			<Field label="Тема" error={errors.subject?.message} {...register('subject')} />
			<TextAreaField label="Сообщение" error={errors.text?.message} {...register('text')} />
			{sent && <Alert tone="success" title="Обращение отправлено">Ответ появится в списке ниже.</Alert>}
			{failure && <Alert tone="danger" title={failure} />}
			<Button type="submit" loading={isLoading}>Отправить обращение</Button>
		</form>
	);
};
