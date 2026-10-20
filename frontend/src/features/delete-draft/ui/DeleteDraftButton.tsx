import { getErrorMessage } from '@shared/api';
import { Alert, Button, Modal } from '@shared/ui';
import { useState } from 'react';

import { useDeleteDraftMutation } from '../api/deleteDraftApi';

interface DeleteDraftButtonProps {
	applicationId: string;
	productName: string;
}

/** Удаление черновика с подтверждением: действие необратимо */
export const DeleteDraftButton = ({ applicationId, productName }: DeleteDraftButtonProps) => {
	const [open, setOpen] = useState(false);
	const [failure, setFailure] = useState<string | null>(null);
	const [deleteDraft, { isLoading }] = useDeleteDraftMutation();

	const handleConfirm = async () => {
		setFailure(null);

		try {
			await deleteDraft(applicationId).unwrap();
			setOpen(false);
		} catch (error) {
			setFailure(getErrorMessage(error, 'Не удалось удалить черновик. Повторите попытку.'));
		}
	};

	return (
		<>
			<Button variant="ghost" compact onClick={() => setOpen(true)}>Удалить черновик</Button>
			{open && (
				<Modal
					title="Удалить черновик?"
					onClose={() => setOpen(false)}
					footer={(
						<>
							<Button variant="secondary" onClick={() => setOpen(false)}>Оставить</Button>
							<Button variant="danger" loading={isLoading} onClick={() => void handleConfirm()}>Удалить</Button>
						</>
					)}
				>
					<p>Черновик «{productName}» будет удалён вместе с введёнными данными. Это нельзя отменить.</p>
					{failure && <Alert tone="danger" title={failure} />}
				</Modal>
			)}
		</>
	);
};
