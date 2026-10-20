import { getErrorMessage } from '@shared/api';
import { Alert, Button, Modal } from '@shared/ui';
import { useState } from 'react';

import { useDeletePlanMutation } from '../api/plansApi';

export const DeletePlan = ({ id, title }: { id: string; title: string }) => {
	const [remove, { isLoading }] = useDeletePlanMutation();
	const [open, setOpen] = useState(false);
	const [failure, setFailure] = useState<string | null>(null);

	const confirm = async () => {
		setFailure(null);

		try {
			await remove(id).unwrap();
			setOpen(false);
		} catch (error) {
			setFailure(getErrorMessage(error, 'Не удалось удалить план'));
		}
	};

	return (
		<>
			<Button variant="ghost" compact onClick={() => setOpen(true)}>{`Удалить план: ${title}`}</Button>
			{open && (
				<Modal
					title="Удалить план?"
					onClose={() => setOpen(false)}
					footer={(
						<>
							<Button variant="ghost" onClick={() => setOpen(false)}>Отмена</Button>
							<Button variant="danger" onClick={() => void confirm()} loading={isLoading}>Удалить</Button>
						</>
					)}
				>
					<p>{`План «${title}» исчезнет из календаря, подбор по нему прекратится.`}</p>
					{failure && <Alert tone="danger" title={failure} />}
				</Modal>
			)}
		</>
	);
};
