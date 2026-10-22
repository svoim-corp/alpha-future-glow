import { useDeletePlanMutation } from '@entities/plan';
import { getErrorMessage } from '@shared/api';
import { Alert, Button, Modal } from '@shared/ui';
import { useState } from 'react';

interface DeletePlanDialogProps {
	id: string;
	title: string;
	onClose: () => void;
}

export const DeletePlanDialog = ({ id, title, onClose }: DeletePlanDialogProps) => {
	const [remove, { isLoading }] = useDeletePlanMutation();
	const [failure, setFailure] = useState<string | null>(null);
	const close = () => { if (!isLoading) onClose(); };
	const confirm = async () => {
		setFailure(null);
		try {
			await remove(id).unwrap();
			onClose();
		} catch (error) {
			setFailure(getErrorMessage(error, 'Не удалось удалить план'));
		}
	};

	return (
		<Modal title="Удалить план?" onClose={close} footer={(
			<>
				<Button variant="ghost" disabled={isLoading} onClick={close}>Отмена</Button>
				<Button variant="danger" onClick={() => void confirm()} loading={isLoading}>Удалить</Button>
			</>
		)}>
			<p>{`План «${title}» исчезнет из календаря, подбор по нему прекратится.`}</p>
			<p>
				Удаление безвозвратное: план не появится в архиве.
				Чтобы сохранить его, отмените удаление и выберите состояние «В архиве».
			</p>
			{failure && <Alert tone="danger" title={failure} />}
		</Modal>
	);
};
