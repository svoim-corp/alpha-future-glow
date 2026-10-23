import { getErrorMessage, getErrorStatus } from '@shared/api';
import { Alert, Button } from '@shared/ui';
import { useState } from 'react';
import { useNavigate } from 'react-router';

import { useCreateApplicationMutation } from '../api/startApplicationApi';
import styles from './StartApplicationButton.module.scss';

interface StartApplicationButtonProps {
	productCode: string;
	amount: number | null;
	term: number | null;
	goal?: string;
	unavailableMessage?: string;
}

/** «Оформить»: создаёт черновик с параметрами из подбора и ведёт в мастер заявки */
export const StartApplicationButton = ({
	productCode, amount, term, goal, unavailableMessage,
}: StartApplicationButtonProps) => {
	const navigate = useNavigate();
	const [create, { isLoading }] = useCreateApplicationMutation();
	const [failure, setFailure] = useState<{ message: string; params: string; unavailable: boolean } | null>(null);
	const params = JSON.stringify({ productCode, amount, term, goal });

	const handleClick = async () => {
		if (amount === null || isLoading) {
			return;
		}

		setFailure(null);

		try {
			const result = await create({ productCode, amount, term, goal }).unwrap();
			await navigate(result.redirect);
		} catch (error) {
			const unsupported = [404, 501].includes(getErrorStatus(error) ?? 0);
			setFailure({
				params,
				unavailable: Boolean(unsupported && unavailableMessage),
				message: unsupported && unavailableMessage
					? unavailableMessage : getErrorMessage(error, 'Не удалось создать заявку. Повторите попытку.'),
			});
		}
	};

	return (
		<div className={styles.root}>
			<Button loading={isLoading} disabled={amount === null} onClick={() => void handleClick()}>Оформить</Button>
			{failure && failure.params === params && (
				<Alert tone={failure.unavailable ? 'warning' : 'danger'} title={failure.message} />
			)}
		</div>
	);
};
