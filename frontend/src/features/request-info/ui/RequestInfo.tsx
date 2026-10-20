import { getErrorMessage } from '@shared/api';
import { Alert, Button, Modal, TextAreaField } from '@shared/ui';
import { useState } from 'react';

import { useRequestInfoMutation } from '../api/requestInfoApi';

/** Модалка «Запрос информации»: общий комментарий и по одному пункту на строку — каждый станет полем ответа клиента */
export const RequestInfo = ({ applicationId }: { applicationId: string }) => {
	const [send, { isLoading }] = useRequestInfoMutation();
	const [open, setOpen] = useState(false);
	const [comment, setComment] = useState('');
	const [itemsText, setItemsText] = useState('');
	const [error, setError] = useState<string | null>(null);

	const items = itemsText.split('\n').map((line) => line.trim()).filter(Boolean);

	const submit = async () => {
		if (!comment.trim() || items.length === 0) {
			setError('Напишите комментарий и хотя бы один вопрос клиенту');

			return;
		}

		setError(null);

		try {
			await send({
				applicationId,
				comment: comment.trim(),
				items: items.map((label, index) => ({ key: `item_${index + 1}`, type: 'TEXT', label })),
			}).unwrap();
			setOpen(false);
			setComment('');
			setItemsText('');
		} catch (caught) {
			setError(getErrorMessage(caught, 'Не удалось отправить запрос'));
		}
	};

	return (
		<>
			<Button variant="secondary" onClick={() => setOpen(true)}>Запросить информацию</Button>
			{open && (
				<Modal
					title="Запрос информации у клиента"
					onClose={() => setOpen(false)}
					footer={(
						<>
							<Button variant="ghost" onClick={() => setOpen(false)}>Отмена</Button>
							<Button onClick={() => void submit()} loading={isLoading}>Отправить запрос</Button>
						</>
					)}
				>
					<TextAreaField label="Комментарий клиенту" value={comment} onChange={(e) => setComment(e.target.value)} />
					<TextAreaField
						label="Вопросы клиенту"
						hint="Каждый вопрос с новой строки"
						value={itemsText}
						onChange={(e) => setItemsText(e.target.value)}
					/>
					{error && <Alert tone="danger" title={error} />}
				</Modal>
			)}
		</>
	);
};
