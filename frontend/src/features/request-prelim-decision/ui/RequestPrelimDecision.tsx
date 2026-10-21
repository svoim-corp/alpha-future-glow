import { type Application } from '@entities/application';
import { getErrorMessage } from '@shared/api';
import { formatMoney, formatRate } from '@shared/lib/format';
import { Alert, Button, CheckboxField, RadioGroup } from '@shared/ui';
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
	{ value: 'REDUCE_AMOUNT', label: 'Уменьшить сумму до лимита' },
	{ value: 'KEEP_AMOUNT', label: 'Оставить запрошенную сумму: решение примет сотрудник банка, возможны иные условия' },
];

/** Согласие на БКИ → предварительное решение (лимит по оборотам в банке или стоп) → выбор, если сумма выше лимита */
export const RequestPrelimDecision = ({ application, onStage2 }: RequestPrelimDecisionProps) => {
	const [giveConsent, consentState] = useGiveBkiConsentMutation();
	const [requestPrelim, prelimState] = useRequestPrelimMutation();
	const [choose, chooseState] = useChoosePrelimMutation();
	const [agreed, setAgreed] = useState(false);
	const [result, setResult] = useState<PrelimResult | null>(null);
	const [choice, setChoice] = useState<string>('REDUCE_AMOUNT');
	const [error, setError] = useState<string | null>(null);

	const authorityOk = application.authority?.verified === true;
	const consented = application.consents.bki?.given === true;
	const stored = application.prelim;
	const limit: PrelimLimit | null = result?.outcome === 'LIMIT' ? result : null;
	const stop = result?.outcome === 'STOP' ? result : null;
	const requested = application.params.amount;
	const bankLimit = limit?.limit ?? stored?.limit ?? null;
	const rate = limit?.prelimRate ?? stored?.prelimRate ?? null;
	const hasLimit = stored?.outcome === 'LIMIT' || limit !== null;
	const needsChoice = limit ? !limit.amountFits : stored?.outcome === 'LIMIT' && stored.amountFits === false
		&& stored.choice === null;

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
			const reduce = choice === 'REDUCE_AMOUNT';
			await choose({
				applicationId: application.id,
				choice: reduce ? 'REDUCE_AMOUNT' : 'KEEP_AMOUNT',
				amount: reduce && bankLimit !== null ? bankLimit : undefined,
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
					<dt>Предварительный лимит</dt>
					<dd>{bankLimit === null ? '—' : formatMoney(bankLimit)}</dd>
				</div>
			</dl>
			<p>Лимит рассчитан по оборотам на ваших счетах в банке, выписки не нужны.</p>
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
					<Alert tone="warning" title="Запрошенная сумма выше лимита">
						{`Вы просили ${formatMoney(requested)}. Выберите, как продолжить.`}
					</Alert>
					<RadioGroup legend="Как продолжить" value={choice} onChange={setChoice} options={CHOICES} />
					<div>
						<Button onClick={() => void confirmChoice()} loading={chooseState.isLoading}>
							Продолжить к полной заявке
						</Button>
					</div>
				</div>
			)}
			{error && <Alert tone="danger" title={error} />}
		</div>
	);
};
