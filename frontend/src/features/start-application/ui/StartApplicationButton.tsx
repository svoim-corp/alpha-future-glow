import { getErrorMessage } from '@shared/api';
import { Alert, Button } from '@shared/ui';
import { useState } from 'react';
import { useNavigate } from 'react-router';

import { useCreateApplicationMutation } from '../api/startApplicationApi';

interface StartApplicationButtonProps {
	productCode: string;
	amount: number | null;
	term: number | null;
	goal?: string;
}

/** «Оформить»: создаёт черновик с параметрами из подбора и ведёт в мастер заявки */
export const StartApplicationButton = ({ productCode, amount, term, goal }: StartApplicationButtonProps) => {
	const navigate = useNavigate();
	const [create, { isLoading }] = useCreateApplicationMutation();
	const [failure, setFailure] = useState<string | null>(null);

	const handleClick = async () => {
		if (amount === null) {
			return;
		}

		setFailure(null);

		try {
			const result = await create({ productCode, amount, term, goal }).unwrap();
			await navigate(result.redirect);
		} catch (error) {
			setFailure(getErrorMessage(error, 'Не удалось создать заявку. Повторите попытку.'));
		}
	};

	return (
		<>
			<Button loading={isLoading} disabled={amount === null} onClick={() => void handleClick()}>Оформить</Button>
			{failure && <Alert tone="danger" title={failure} />}
		</>
	);
};
