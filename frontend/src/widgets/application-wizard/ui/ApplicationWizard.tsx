import { type Application, CompanyProfileView, useGetApplicationQuery } from '@entities/application';
import { PaymentEstimate } from '@features/calculate-payment';
import { EditDraftParams, type ParamsValues } from '@features/edit-draft-params';
import { FillFullApplication } from '@features/fill-full-application';
import { RequestPrelimDecision } from '@features/request-prelim-decision';
import { VerifyCompany } from '@features/verify-company';
import { paths } from '@shared/config';
import { formatMoney, formatRate, formatTerm } from '@shared/lib/format';
import { AsyncBoundary, Button, Card, Steps } from '@shared/ui';
import { useState } from 'react';
import { Navigate, useNavigate } from 'react-router';

import styles from './ApplicationWizard.module.scss';

const STEP_LABELS = ['Этап 1. Предварительное решение', 'Этап 2. Полная заявка'];

const SectionTitle = ({ id, number, children }: { id: string; number: number; children: string }) => (
	<h2 id={id} className={styles.title}>
		<span className={styles.num} aria-hidden="true">{number}</span>
		{children}
	</h2>
);

const toNumber = (text: string): number | null => (/^\d+$/u.test(text) && Number(text) > 0 ? Number(text) : null);

/** Сводка по заявке рядом с формой: что выбрано (обновляется вместе с расчётом платежа) и какой лимит назвал банк */
const Summary = ({ application, draft }: { application: Application; draft: ParamsValues | null }) => {
	/* после предварительного решения параметры зафиксированы: верим только сохранённым значениям */
	const typed = application.locks.paramsLocked ? null : draft;
	const amount = toNumber(typed?.amount ?? '') ?? application.params.amount;
	const term = typed ? toNumber(typed.term ?? '') ?? application.params.term : application.params.term;

	return (
		<aside className={styles.summary} aria-label="Сводка по заявке">
			<span className={styles.eyebrow}>ВАША ЗАЯВКА</span>
			<h2 className={styles.summaryTitle}>{application.product.name}</h2>
			<dl className={styles.facts}>
				<div>
					<dt>Сумма</dt>
					<dd>{formatMoney(amount)}</dd>
				</div>
				{term !== null && (
					<div>
						<dt>Срок</dt>
						<dd>{formatTerm(term)}</dd>
					</div>
				)}
				{application.prelim?.outcome === 'LIMIT' && application.prelim.limit !== null && (
					<div>
						<dt>Лимит банка</dt>
						<dd>{formatMoney(application.prelim.limit)}</dd>
					</div>
				)}
				{application.prelim?.prelimRate !== undefined && application.prelim.prelimRate !== null && (
					<div>
						<dt>Ставка</dt>
						<dd>{formatRate(application.prelim.prelimRate)}</dd>
					</div>
				)}
			</dl>
		</aside>
	);
};

export const ApplicationWizard = ({ applicationId }: { applicationId: string }) => {
	const { data: application, isLoading, isError, refetch } = useGetApplicationQuery(applicationId);
	const navigate = useNavigate();
	const [stage2Shown, setStage2Shown] = useState<boolean | null>(null);
	const [openedAtStage2, setOpenedAtStage2] = useState<boolean | null>(null);
	const [draft, setDraft] = useState<ParamsValues | null>(null);

	/* Этап при открытии запоминаем один раз: после запроса решения банк переводит заявку на этап 2,
	   но клиент должен остаться на первом и увидеть само решение */
	if (application && openedAtStage2 === null) setOpenedAtStage2(application.stage === 2);

	if (application && application.status !== 'DRAFT') {
		return <Navigate to={paths.application(applicationId)} replace />;
	}

	const onStage2 = application?.stage === 2 || application?.locks.stage2Available === true;
	const showStage2 = stage2Shown ?? openedAtStage2 === true;

	return (
		<AsyncBoundary
			isLoading={isLoading}
			isError={isError}
			onRetry={() => void refetch()}
			loadingLabel="Загружаем заявку"
			errorText="Не удалось загрузить заявку"
		>
			{application && (
				<div className={styles.wizard}>
					<Steps label="Этапы заявки" labels={STEP_LABELS} current={showStage2 ? 1 : 0} />
					<div className={styles.layout}>
						<div className={styles.main}>
							{showStage2 ? (
								<>
									<Card className={styles.card} aria-labelledby="w-profile">
										<SectionTitle id="w-profile" number={1}>Данные вашей компании</SectionTitle>
										<CompanyProfileView applicationId={application.id} />
									</Card>
									<Card className={styles.card} aria-labelledby="w-full">
										<SectionTitle id="w-full" number={2}>Полная заявка</SectionTitle>
										<FillFullApplication
											application={application}
											onSubmitted={(redirect) => void navigate(redirect)}
										/>
									</Card>
									<div className={styles.actions}>
										<Button variant="ghost" onClick={() => setStage2Shown(false)}>
											Вернуться к этапу 1
										</Button>
									</div>
								</>
							) : (
								<>
									<Card className={styles.card} aria-labelledby="w-company">
										<SectionTitle id="w-company" number={1}>Компания</SectionTitle>
										<VerifyCompany application={application} />
									</Card>
									<Card className={styles.card} aria-labelledby="w-params">
										<SectionTitle id="w-params" number={2}>Параметры кредита</SectionTitle>
										<EditDraftParams
											key={application.prelim?.choice ? `chosen-${application.params.amount}` : 'form'}
											application={application}
											onChange={setDraft}
											estimate={(
												<PaymentEstimate
													params={{
														productCode: application.product.code,
														amount: toNumber(draft?.amount ?? `${application.params.amount}`) ?? 0,
														term: toNumber(draft?.term ?? `${application.params.term ?? ''}`) ?? 0,
														graceMonths: toNumber(draft?.graceMonths ?? '') ?? application.params.graceMonths,
													}}
												/>
											)}
										/>
									</Card>
									<Card className={styles.card} aria-labelledby="w-prelim">
										<SectionTitle id="w-prelim" number={3}>Предварительное решение</SectionTitle>
										<RequestPrelimDecision
											application={application}
											onChosen={() => {
												setDraft(null);
												setStage2Shown(true);
											}}
										/>
									</Card>
									{onStage2 && (
										<div className={styles.actions}>
											<Button onClick={() => setStage2Shown(true)}>
												Перейти к полной заявке
											</Button>
										</div>
									)}
								</>
							)}
						</div>
						<Summary application={application} draft={draft} />
					</div>
				</div>
			)}
		</AsyncBoundary>
	);
};
