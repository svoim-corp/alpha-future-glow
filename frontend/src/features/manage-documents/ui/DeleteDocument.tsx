import { getErrorMessage } from '@shared/api';
import { Alert, Button, Icon, Modal } from '@shared/ui';
import { useState } from 'react';

import { useDeleteDocumentMutation } from '../api/deleteApi';

export const DeleteDocument = ({ id, name }: { id: string; name: string }) => {
	const [remove, { isLoading }] = useDeleteDocumentMutation();
	const [open, setOpen] = useState(false);
	const [failure, setFailure] = useState<string | null>(null);

	const confirm = async () => {
		setFailure(null);

		try {
			await remove(id).unwrap();
			setOpen(false);
		} catch (error) {
			setFailure(getErrorMessage(error, 'Не удалось удалить файл'));
		}
	};

	return (
		<>
			<Button variant="ghost" compact aria-label={`Удалить: ${name}`} onClick={() => setOpen(true)}>
				<Icon name="close" size={16} />
			</Button>
			{open && (
				<Modal
					title="Удалить документ?"
					onClose={() => setOpen(false)}
					footer={(
						<>
							<Button variant="ghost" onClick={() => setOpen(false)}>Отмена</Button>
							<Button variant="danger" onClick={() => void confirm()} loading={isLoading}>Удалить</Button>
						</>
					)}
				>
					<p>{`«${name}» исчезнет из библиотеки. Копии, уже отправленные в банк с заявкой, сохранятся.`}</p>
					{failure && <Alert tone="danger" title={failure} />}
				</Modal>
			)}
		</>
	);
};
