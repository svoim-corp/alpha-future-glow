import { type BankQueueItem, useAssignApplicationMutation } from '@entities/application';
import { getApiError } from '@shared/api';
import { Alert, Button } from '@shared/ui';
import { useState } from 'react';

/** «Взять в работу»: если заявку уже взял другой сотрудник, сервер ответит 409 — показываем кто */
export const AssignApplication = ({ item }: { item: Pick<BankQueueItem, 'id' | 'number' | 'assignee'> }) => {
	const [assign, { isLoading }] = useAssignApplicationMutation();
	const [failure, setFailure] = useState<string | null>(null);

	if (item.assignee) {
		return <span>{item.assignee.fullName}</span>;
	}

	const onClick = async () => {
		setFailure(null);

		try {
			await assign(item.id).unwrap();
		} catch (error) {
			setFailure(getApiError(error)?.code === 'CONFLICT'
				? 'Заявку уже взял другой сотрудник'
				: getApiError(error)?.message ?? 'Не удалось взять заявку в работу');
		}
	};

	return (
		<>
			<Button
				variant="secondary"
				compact
				loading={isLoading}
				aria-label={`Взять в работу ${item.number}`}
				onClick={() => void onClick()}
			>
				Взять в работу
			</Button>
			{failure && <Alert tone="danger" title={failure} />}
		</>
	);
};
