import { getErrorMessage } from '@shared/api';
import { Alert, Button, Field, Modal, TextAreaField } from '@shared/ui';
import { useState } from 'react';

import { type DecisionArgs, useMakeDecisionMutation } from '../api/decisionApi';
import styles from './MakeDecision.module.scss';

interface MakeDecisionProps {
	applicationId: string;
	requested: { productCode: string; amount: number; term: number };
	/** Ставка из рекомендации скоринга или предварительной проверки: бэкенд требует её и для одобрения, и для иных условий */
	defaultRate?: number;
}

type Mode = 'COUNTER' | 'REJECT' | null;

const toNumber = (text: string): number | null => {
	const value = text.trim();

	return /^\d+$/u.test(value) && Number(value) > 0 ? Number(value) : null;
};

/** Решение сотрудника: одобрить как запрошено, предложить иные условия (модалка) или отказать с причиной */
export const MakeDecision = ({ applicationId, requested, defaultRate }: MakeDecisionProps) => {
	const [decide, { isLoading }] = useMakeDecisionMutation();
	const [mode, setMode] = useState<Mode>(null);
	const [amount, setAmount] = useState(String(requested.amount));
	const [term, setTerm] = useState(String(requested.term));
	const [rate, setRate] = useState(defaultRate ? String(defaultRate) : '');
	const [reason, setReason] = useState('');
	const [error, setError] = useState<string | null>(null);
	const [done, setDone] = useState<string | null>(null);
	// один ключ на одно решение: повторный клик не создаст второе
	const [key] = useState(() => crypto.randomUUID());

	const send = async (body: DecisionArgs['body']) => {
		setError(null);

		try {
			const result = await decide({ applicationId, idempotencyKey: key, body }).unwrap();
			setDone(result.status);
			setMode(null);
		} catch (caught) {
			setError(getErrorMessage(caught, 'Не удалось сохранить решение'));
		}
	};

	const submitCounter = () => {
		const parsedAmount = toNumber(amount);
		const parsedTerm = toNumber(term);

		const parsedRate = Number(rate.replace(',', '.'));

		if (!parsedAmount || !parsedTerm || !reason.trim()) {
			setError('Укажите сумму, срок и объяснение для клиента');

			return;
		}

		if (!Number.isFinite(parsedRate) || parsedRate <= 0) {
			setError('Укажите ставку, % годовых');

			return;
		}

		void send({
			type: 'COUNTER',
			amount: parsedAmount,
			term: parsedTerm,
			productCode: requested.productCode,
			rate: parsedRate,
			reason: reason.trim(),
		});
	};

	const submitReject = () => {
		if (!reason.trim()) {
			setError('Укажите причину отказа — клиент её увидит');

			return;
		}

		void send({ type: 'REJECT', reason: reason.trim() });
	};

	const close = () => {
		setMode(null);
		setError(null);
	};

	if (done) {
		return <Alert tone="success" title="Решение сохранено">{`Новый статус заявки: ${done}`}</Alert>;
	}

	return (
		<div className={styles.root}>
			{error && !mode && <Alert tone="danger" title={error} />}
			{!defaultRate && (
				<Alert tone="info" title="Ставка не рассчитана">
					Одобрить как запрошено нельзя: укажите ставку через «Иные условия».
				</Alert>
			)}
			<div className={styles.actions}>
				<Button
					loading={isLoading}
					disabled={!defaultRate}
					onClick={() => void send({ type: 'APPROVE', ...requested, rate: defaultRate })}
				>
					Одобрить как запрошено
				</Button>
				<Button variant="danger" onClick={() => setMode('REJECT')}>Отказать</Button>
				<Button variant="secondary" onClick={() => setMode('COUNTER')}>Иные условия</Button>
			</div>
			{mode === 'COUNTER' && (
				<Modal
					title="Иные условия"
					onClose={close}
					footer={(
						<>
							<Button variant="ghost" onClick={close}>Отмена</Button>
							<Button onClick={submitCounter} loading={isLoading}>Предложить клиенту</Button>
						</>
					)}
				>
					<Field label="Сумма, ₽" inputMode="numeric" value={amount} onChange={(e) => setAmount(e.target.value)} />
					<Field label="Срок, мес." inputMode="numeric" value={term} onChange={(e) => setTerm(e.target.value)} />
					<Field label="Ставка, % годовых" inputMode="decimal" value={rate} onChange={(e) => setRate(e.target.value)} />
					<TextAreaField label="Объяснение для клиента" value={reason} onChange={(e) => setReason(e.target.value)} />
					{error && <Alert tone="danger" title={error} />}
				</Modal>
			)}
			{mode === 'REJECT' && (
				<Modal
					title="Отказать в заявке"
					onClose={close}
					footer={(
						<>
							<Button variant="ghost" onClick={close}>Отмена</Button>
							<Button variant="danger" onClick={submitReject} loading={isLoading}>Отказать</Button>
						</>
					)}
				>
					<TextAreaField label="Причина отказа" value={reason} onChange={(e) => setReason(e.target.value)} />
					{error && <Alert tone="danger" title={error} />}
				</Modal>
			)}
		</div>
	);
};
