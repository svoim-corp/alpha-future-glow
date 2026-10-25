import { type Contract } from '@entities/contract';
import { getErrorMessage } from '@shared/api';
import { Alert, Button, ButtonLink, Modal } from '@shared/ui';
import { useState } from 'react';

import { type DisburseResult, useDisburseLoanMutation, useSignContractMutation } from '../api/signApi';
import styles from './SignContract.module.scss';

/** Имитация подписи без SMS, сразу за ней — имитация выдачи. Ключ идемпотентности один на действие — повтор безопасен */
export const SignContract = ({ applicationId, contract }: { applicationId: string; contract: Contract }) => {
	const [sign, signState] = useSignContractMutation();
	const [disburse, disburseState] = useDisburseLoanMutation();
	const [signKey] = useState(() => crypto.randomUUID());
	const [disburseKey] = useState(() => crypto.randomUUID());
	const [confirming, setConfirming] = useState(false);
	const [failure, setFailure] = useState<string | null>(null);
	const [issued, setIssued] = useState<DisburseResult | null>(null);

	const onDisburse = async () => {
		setFailure(null);

		try {
			setIssued(await disburse({ applicationId, idempotencyKey: disburseKey }).unwrap());
		} catch (error) {
			setFailure(getErrorMessage(error, 'Не удалось выдать кредит. Повторите попытку.'));
		}
	};

	const onSign = async () => {
		setFailure(null);

		try {
			await sign({ applicationId, idempotencyKey: signKey }).unwrap();
			setConfirming(false);
		} catch (error) {
			setFailure(getErrorMessage(error, 'Не удалось подписать договор. Повторите попытку.'));

			return;
		}

		// Отдельного шага «получить деньги» нет: после подписи кредит выдаётся сразу
		await onDisburse();
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
					<Alert tone="success" title="Договор подписан">
						{disburseState.isLoading ? 'Зачисляем деньги на ваш счёт…' : 'Деньги пока не зачислены.'}
					</Alert>
					{!disburseState.isLoading && (
						<div>
							<Button onClick={() => void onDisburse()}>Повторить зачисление</Button>
						</div>
					)}
				</>
			) : (
				<div>
					<Button onClick={() => setConfirming(true)}>Подписать договор</Button>
				</div>
			)}
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
