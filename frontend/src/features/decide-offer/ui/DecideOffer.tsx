import { getErrorMessage } from '@shared/api';
import { Alert, Button, Modal, TextAreaField } from '@shared/ui';
import { useState } from 'react';

import { useAcceptOfferMutation, useDeclineOfferMutation } from '../api/offerDecisionApi';
import styles from './DecideOffer.module.scss';

interface DecideOfferProps {
	applicationId: string;
	/** Куда ведёт бэк после принятия; страница заявки сама покажет подписание, поэтому необязательно */
	onAccepted?: (redirect: string) => void;
}

/** Принять условия (→ подписание) или отказаться с необязательной причиной (→ заявка закрывается) */
export const DecideOffer = ({ applicationId, onAccepted }: DecideOfferProps) => {
	const [accept, acceptState] = useAcceptOfferMutation();
	const [decline, declineState] = useDeclineOfferMutation();
	const [confirming, setConfirming] = useState(false);
	const [reason, setReason] = useState('');
	const [failure, setFailure] = useState<string | null>(null);

	const onAccept = async () => {
		setFailure(null);

		try {
			const result = await accept({ applicationId }).unwrap();
			onAccepted?.(result.redirect);
		} catch (error) {
			setFailure(getErrorMessage(error, 'Не удалось принять условия. Повторите попытку.'));
		}
	};

	const onDecline = async () => {
		setFailure(null);

		try {
			await decline({ applicationId, reason: reason.trim() }).unwrap();
			setConfirming(false);
		} catch (error) {
			setFailure(getErrorMessage(error, 'Не удалось отказаться от предложения. Повторите попытку.'));
		}
	};

	return (
		<div>
			{failure && !confirming && <Alert tone="danger" title={failure} />}
			<div className={styles.actions}>
				<Button onClick={() => void onAccept()} loading={acceptState.isLoading}>Принять условия</Button>
				<Button variant="secondary" onClick={() => setConfirming(true)}>Отказаться</Button>
			</div>
			{confirming && (
				<Modal
					title="Отказаться от предложения?"
					onClose={() => setConfirming(false)}
					footer={(
						<>
							<Button variant="ghost" onClick={() => setConfirming(false)}>Вернуться</Button>
							<Button variant="danger" onClick={() => void onDecline()} loading={declineState.isLoading}>
								Отказаться
							</Button>
						</>
					)}
				>
					<p>Заявка будет закрыта. Новую заявку можно оформить в любой момент.</p>
					<TextAreaField
						label="Причина (необязательно)"
						value={reason}
						onChange={(event) => setReason(event.target.value)}
					/>
					{failure && <Alert tone="danger" title={failure} />}
				</Modal>
			)}
		</div>
	);
};
