import { zodResolver } from '@hookform/resolvers/zod';
import { getErrorMessage } from '@shared/api';
import { Alert, Button, TextAreaField } from '@shared/ui';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';

import { useCreateConversationMutation, useSendSupportMessageMutation } from '../api/chatApi';
import styles from './SendSupportMessage.module.scss';

const schema = z.object({ text: z.string().trim().min(1, 'Напишите вопрос').max(4000, 'Не больше 4000 символов') });

interface SendSupportMessageProps {
	conversationId?: string;
	applicationId?: string;
	busy: boolean;
	onConversationCreated: (id: string) => void;
}

export const SendSupportMessage = ({
	conversationId, applicationId, busy, onConversationCreated,
}: SendSupportMessageProps) => {
	const [create, creation] = useCreateConversationMutation();
	const [send, sending] = useSendSupportMessageMutation();
	const [failure, setFailure] = useState<string | null>(null);
	const [request, setRequest] = useState<{ conversationId: string; text: string; key: string } | null>(null);
	const { register, handleSubmit, reset, formState: { errors } } = useForm<z.infer<typeof schema>>({
		resolver: zodResolver(schema),
	});
	const submit = handleSubmit(async ({ text }) => {
		setFailure(null);
		try {
			let id = conversationId;
			if (!id) {
				const conversation = await create({ applicationId }).unwrap();
				id = conversation.id;
				onConversationCreated(id);
			}
			const previous = request;
			const args = previous?.text === text && previous.conversationId === id
				? previous : { conversationId: id, text, key: crypto.randomUUID() };
			setRequest(args);
			await send(args).unwrap();
			setRequest(null);
			reset({ text: '' });
		} catch (error) {
			setFailure(getErrorMessage(error, 'Не удалось отправить вопрос. Текст сохранён, попробуйте ещё раз.'));
		}
	});

	return (
		<form className={styles.form} onSubmit={(event) => void submit(event)} noValidate>
			<TextAreaField label="Ваш вопрос" placeholder="Например: когда мой ближайший платёж?" rows={3} maxLength={4000} error={errors.text?.message} {...register('text')} />
			{failure && <Alert tone="danger" title={failure} />}
			<div className={styles.footer}>
				<span>Создание черновика и обращения — только после подтверждения.</span>
				<Button type="submit" disabled={busy} loading={creation.isLoading || sending.isLoading}>Отправить вопрос</Button>
			</div>
		</form>
	);
};
