import { CompanyProfileView, useGetApplicationQuery } from '@entities/application';
import { PaymentEstimate } from '@features/calculate-payment';
import { EditDraftParams, type ParamsValues } from '@features/edit-draft-params';
import { FillFullApplication } from '@features/fill-full-application';
import { RequestPrelimDecision } from '@features/request-prelim-decision';
import { UploadStatement } from '@features/upload-statement';
import { VerifyCompany } from '@features/verify-company';
import { paths } from '@shared/config';
import { AsyncBoundary, Button, Card, Steps } from '@shared/ui';
import { useState } from 'react';
import { Navigate, useNavigate } from 'react-router';

import styles from './ApplicationWizard.module.scss';

const STEP_LABELS = ['Этап 1. Предварительное решение', 'Этап 2. Полная заявка'];

const toNumber = (text: string): number | null => (/^\d+$/u.test(text) && Number(text) > 0 ? Number(text) : null);

export const ApplicationWizard = ({ applicationId }: { applicationId: string }) => {
	const { data: application, isLoading, isError, refetch } = useGetApplicationQuery(applicationId);
	const navigate = useNavigate();
	const [stage2Shown, setStage2Shown] = useState<boolean | null>(null);
	const [draft, setDraft] = useState<ParamsValues | null>(null);

	if (application && application.status !== 'DRAFT') {
		return <Navigate to={paths.application(applicationId)} replace />;
	}

	const onStage2 = application?.stage === 2 || application?.locks.stage2Available === true;
	const showStage2 = stage2Shown ?? application?.stage === 2;

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
					{showStage2 ? (
						<>
							<Card className={styles.card} aria-labelledby="w-profile">
								<h2 id="w-profile" className={styles.title}>Данные вашей компании</h2>
								<CompanyProfileView applicationId={application.id} />
							</Card>
							<Card className={styles.card} aria-labelledby="w-statement">
								<h2 id="w-statement" className={styles.title}>Выписка по счёту</h2>
								<UploadStatement applicationId={application.id} />
							</Card>
							<Card className={styles.card} aria-labelledby="w-full">
								<h2 id="w-full" className={styles.title}>Полная заявка</h2>
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
								<h2 id="w-company" className={styles.title}>Компания</h2>
								<VerifyCompany application={application} />
							</Card>
							<Card className={styles.card} aria-labelledby="w-params">
								<h2 id="w-params" className={styles.title}>Параметры кредита</h2>
								<EditDraftParams
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
								<h2 id="w-prelim" className={styles.title}>Предварительное решение</h2>
								<RequestPrelimDecision
									application={application}
									onStage2={() => setStage2Shown(true)}
								/>
							</Card>
							{onStage2 && (
								<div className={styles.actions}>
									<Button onClick={() => setStage2Shown(true)}>Перейти к полной заявке</Button>
								</div>
							)}
						</>
					)}
				</div>
			)}
		</AsyncBoundary>
	);
};
