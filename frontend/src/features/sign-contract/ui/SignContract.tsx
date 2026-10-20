import { type Contract } from '@entities/contract';
import { getErrorMessage } from '@shared/api';
import { paths } from '@shared/config';
import { Alert, Button, ButtonLink, Modal } from '@shared/ui';
import { useState } from 'react';

import { type DisburseResult, useDisburseLoanMutation, useSignContractMutation } from '../api/signApi';
import styles from './SignContract.module.scss';

/** Имитация подписи без SMS, затем имитация выдачи. Ключ идемпотентности один на действие — повтор безопасен */
export const SignContract = ({ applicationId, contract }: { applicationId: string; contract: Contract }) => {
	const [sign, signState] = useSignContractMutation();
	const [disburse, disburseState] = useDisburseLoanMutation();
	const [signKey] = useState(() => crypto.randomUUID());
	const [disburseKey] = useState(() => crypto.randomUUID());
	const [confirming, setConfirming] = useState(false);
	const [failure, setFailure] = useState<string | null>(null);
	const [issued, setIssued] = useState<DisburseResult | null>(null);

	const onSign = async () => {
		setFailure(null);

		try {
			await sign({ applicationId, idempotencyKey: signKey }).unwrap();
			setConfirming(false);
		} catch (error) {
			setFailure(getErrorMessage(error, 'Не удалось подписать договор. Повторите попытку.'));
		}
	};

	const onDisburse = async () => {
		setFailure(null);

		try {
			setIssued(await disburse({ applicationId, idempotencyKey: disburseKey }).unwrap());
		} catch (error) {
			setFailure(getErrorMessage(error, 'Не удалось выдать кредит. Повторите попытку.'));
		}
	};

	if (issued || contract.disbursed) {
		return (
			<div className={styles.block}>
				<Alert tone="success" title="Кредит выдан">
					{issued
						? `Деньги зачислены на счёт ${issued.accountMasked}. График платежей уже готов.`
						: 'Деньги зачислены на ваш счёт. График платежей уже готов.'}
				</Alert>
				<div><ButtonLink to="/credits">Перейти в «Мои кредиты»</ButtonLink></div>
			</div>
		);
	}

	return (
		<div className={styles.block}>
			{failure && !confirming && <Alert tone="danger" title={failure} />}
			{contract.signed ? (
				<>
					<Alert tone="success" title="Договор подписан">Остался последний шаг — получить деньги.</Alert>
					<div>
						<Button onClick={() => void onDisburse()} loading={disburseState.isLoading}>
							Получить деньги
						</Button>
					</div>
				</>
			) : (
				<div>
					<Button onClick={() => setConfirming(true)}>Подписать договор</Button>
				</div>
			)}
			<div><ButtonLink variant="ghost" to={paths.application(applicationId)}>К заявке</ButtonLink></div>
			{confirming && (
				<Modal
					title="Подтверждение подписи"
					onClose={() => setConfirming(false)}
					footer={(
						<>
							<Button variant="ghost" onClick={() => setConfirming(false)}>Отмена</Button>
							<Button onClick={() => void onSign()} loading={signState.isLoading}>Подтверждаю</Button>
						</>
					)}
				>
					<p>Вы подписываете кредитный договор на указанных условиях. В демонстрации код из SMS не нужен.</p>
					{failure && <Alert tone="danger" title={failure} />}
				</Modal>
			)}
		</div>
	);
};
