import { useRefreshChecksMutation } from '@entities/application';
import { getErrorMessage } from '@shared/api';
import { Alert, Button } from '@shared/ui';
import { useState } from 'react';

/** Перезапрос внешних проверок: свежие данные сразу попадают в список через инвалидацию кэша */
export const RefreshChecks = ({ applicationId }: { applicationId: string }) => {
	const [refresh, { isLoading }] = useRefreshChecksMutation();
	const [failure, setFailure] = useState<string | null>(null);

	const onClick = async () => {
		setFailure(null);

		try {
			await refresh({ id: applicationId }).unwrap();
		} catch (error) {
			setFailure(getErrorMessage(error, 'Не удалось обновить проверки'));
		}
	};

	return (
		<div>
			<Button variant="secondary" compact loading={isLoading} onClick={() => void onClick()}>
				Обновить проверки
			</Button>
			{failure && <Alert tone="danger" title={failure} />}
		</div>
	);
};
