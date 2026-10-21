import { type Application } from '@entities/application';
import { getErrorMessage } from '@shared/api';
import { formatMoney, formatRate } from '@shared/lib/format';
import { Alert, Button, CheckboxField, Field, RadioGroup } from '@shared/ui';
import { useState } from 'react';

import {
	type PrelimLimit,
	type PrelimResult,
	useChoosePrelimMutation,
	useGiveBkiConsentMutation,
	useRequestPrelimMutation,
} from '../api/prelimApi';
import styles from './RequestPrelimDecision.module.scss';

interface RequestPrelimDecisionProps {
	application: Application;
	/** Нужен числовой ввод суммы выше формы параметров: после выбора «уменьшить» параметры обновятся из заявки */
	onStage2?: () => void;
}

const CHECK_TONE = { OK: 'Ок', WARN: 'Внимание', FAIL: 'Проблема' } as const;

const CHOICES = [
	{ value: 'STATEMENT', label: 'Подтвердить обороты выпиской по счёту' },
	{ value: 'REDUCE_AMOUNT', label: 'Уменьшить сумму до лимита без выписки' },
];

/** Согласие на БКИ → предварительное решение (два лимита или стоп) → выбор, если сумма выше лимита без выписки */
export const RequestPrelimDecision = ({ application, onStage2 }: RequestPrelimDecisionProps) => {
	const [giveConsent, consentState] = useGiveBkiConsentMutation();
	const [requestPrelim, prelimState] = useRequestPrelimMutation();
	const [choose, chooseState] = useChoosePrelimMutation();
	const [agreed, setAgreed] = useState(false);
	const [result, setResult] = useState<PrelimResult | null>(null);
	const [choice, setChoice] = useState<string>('STATEMENT');
	const [amount, setAmount] = useState('');
	const [error, setError] = useState<string | null>(null);

	const authorityOk = application.authority?.verified === true;
	const consented = application.consents.bki?.given === true;
	const stored = application.prelim;
	const limit: PrelimLimit | null = result?.outcome === 'LIMIT' ? result : null;
	const stop = result?.outcome === 'STOP' ? result : null;
	const requested = application.params.amount;
	const limitNoStatement = limit?.limitNoStatement ?? stored?.limitNoStatement ?? null;
	const limitWithStatement = limit?.limitWithStatement ?? stored?.limitWithStatement ?? null;
	const rate = limit?.prelimRate ?? stored?.prelimRate ?? null;
	const hasLimit = stored?.outcome === 'LIMIT' || limit !== null;
	const needsChoice = limit ? limit.needsChoice : stored?.outcome === 'LIMIT' && stored.choice === null
		&& limitNoStatement !== null && requested > limitNoStatement;

	const run = async () => {
		setError(null);

		try {
			if (!consented) {
				await giveConsent(application.id).unwrap();
			}

			setResult(await requestPrelim(application.id).unwrap());
		} catch (caught) {
			setError(getErrorMessage(caught, 'Не удалось получить предварительное решение. Повторите попытку.'));
		}
	};

	const confirmChoice = async () => {
		setError(null);

		try {
			const parsed = Number(amount);
			await choose({
				applicationId: application.id,
				choice: choice === 'REDUCE_AMOUNT' ? 'REDUCE_AMOUNT' : 'STATEMENT',
				amount: choice === 'REDUCE_AMOUNT' ? parsed : undefined,
			}).unwrap();
			onStage2?.();
		} catch (caught) {
			setError(getErrorMessage(caught, 'Не удалось сохранить выбор. Повторите попытку.'));
		}
	};

	if (stop) {
		return (
			<div className={styles.block}>
				<Alert tone="danger" title="Мы не можем одобрить заявку сейчас">{stop.stopText}</Alert>
				<ul className={styles.list}>
					{stop.whatToDo.map((item) => <li key={item}>{item}</li>)}
				</ul>
			</div>
		);
	}

	if (!hasLimit) {
		return (
			<div className={styles.block}>
				{!authorityOk && (
					<Alert tone="info" title="Сначала подтвердите полномочия">
						Согласие на запрос в БКИ доступно после проверки руководителя или доверенности.
					</Alert>
				)}
				<CheckboxField
					label="Согласен на запрос моей кредитной истории в бюро кредитных историй"
					checked={consented || agreed}
					disabled={consented || !authorityOk}
					onChange={(event) => setAgreed(event.target.checked)}
				/>
				<div>
					<Button
						onClick={() => void run()}
						loading={prelimState.isLoading || consentState.isLoading}
						disabled={!authorityOk || !(consented || agreed)}
					>
						Получить предварительное решение
					</Button>
				</div>
				{error && <Alert tone="danger" title={error} />}
			</div>
		);
	}

	return (
		<div className={styles.block}>
			<dl className={styles.limits}>
				<div className={styles.limit}>
					<dt>Лимит без выписки</dt>
					<dd>{limitNoStatement === null ? '—' : formatMoney(limitNoStatement)}</dd>
				</div>
				<div className={styles.limit}>
					<dt>Лимит с выпиской</dt>
					<dd>{limitWithStatement === null ? '—' : formatMoney(limitWithStatement)}</dd>
				</div>
			</dl>
			{rate !== null && (
				<p>
					{`Предварительная ставка ${formatRate(rate)}`}
					{limit ? `, платёж ${formatMoney(limit.prelimPayment)} в месяц` : ''}
				</p>
			)}
			{limit && (
				<ul className={styles.list} aria-label="Результаты проверок">
					{limit.checks.map((check) => (
						<li key={check.source}>{`${CHECK_TONE[check.status]}: ${check.label}`}</li>
					))}
				</ul>
			)}
			{needsChoice && (
				<div className={styles.choice}>
					<Alert tone="warning" title="Запрошенная сумма выше лимита без выписки">
						{`Вы просили ${formatMoney(requested)}. Выберите, как продолжить.`}
					</Alert>
					<RadioGroup legend="Как продолжить" value={choice} onChange={setChoice} options={CHOICES} />
					{choice === 'REDUCE_AMOUNT' && (
						<Field
							label="Новая сумма, ₽"
							inputMode="numeric"
							value={amount}
							onChange={(event) => setAmount(event.target.value)}
							hint={limitNoStatement === null ? undefined : `Не больше ${formatMoney(limitNoStatement)}`}
						/>
					)}
					<div>
						<Button
							onClick={() => void confirmChoice()}
							loading={chooseState.isLoading}
							disabled={choice === 'REDUCE_AMOUNT' && !/^\d+$/u.test(amount)}
						>
							Продолжить к полной заявке
						</Button>
					</div>
				</div>
			)}
			{error && <Alert tone="danger" title={error} />}
		</div>
	);
};
