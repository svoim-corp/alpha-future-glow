import {
	ApplicantView,
	BANK_STATUS_LABEL,
	ChecksView,
	CompletenessView,
	HistoryView,
	InfoRequestsView,
	ScoreView,
	StatusBadge,
	useGetBankApplicationQuery,
	useGetBankHistoryQuery,
	useGetBankInfoRequestsQuery,
	useGetChecksQuery,
	useGetCompletenessQuery,
	useGetScoreQuery,
} from '@entities/application';
import { MakeDecision } from '@features/make-decision';
import { RefreshChecks } from '@features/refresh-checks';
import { RequestInfo } from '@features/request-info';
import { AsyncBoundary, Card } from '@shared/ui';
import { type ReactNode } from 'react';

import styles from './BankReview.module.scss';

interface SectionProps {
	id: string;
	title: string;
	isLoading: boolean;
	isError: boolean;
	onRetry: () => void;
	children: ReactNode;
	extra?: ReactNode;
}

/** Секция грузится независимо: сбой скоринга не прячет анкету и историю */
const Section = ({ id, title, isLoading, isError, onRetry, children, extra }: SectionProps) => (
	<Card className={styles.card} aria-labelledby={id}>
		<div className={styles.head}>
			<h2 id={id} className={styles.title}>{title}</h2>
			{extra}
		</div>
		<AsyncBoundary
			isLoading={isLoading}
			isError={isError}
			onRetry={onRetry}
			loadingLabel={`Загружаем: ${title}`}
			errorText={`Не удалось загрузить раздел «${title}»`}
		>
			{children}
		</AsyncBoundary>
	</Card>
);

export const BankReview = ({ applicationId }: { applicationId: string }) => {
	const application = useGetBankApplicationQuery(applicationId);
	const completeness = useGetCompletenessQuery(applicationId);
	const checks = useGetChecksQuery(applicationId);
	const score = useGetScoreQuery(applicationId);
	const history = useGetBankHistoryQuery(applicationId);
	const infoRequests = useGetBankInfoRequestsQuery(applicationId);
	const data = application.data;
	const decidable = data?.status === 'ON_DECISION' || data?.status === 'SCORING' || data?.status === 'DATA_CHECK';

	return (
		<div className={styles.review}>
			<div className={styles.main}>
				<Section
					id="r-applicant"
					title="Анкета"
					isLoading={application.isLoading}
					isError={application.isError}
					onRetry={() => void application.refetch()}
					extra={data && <StatusBadge status={data.status} label={BANK_STATUS_LABEL[data.status]} />}
				>
					{data && <ApplicantView data={data} />}
				</Section>
				<Section
					id="r-completeness"
					title="Полнота и корректность"
					isLoading={completeness.isLoading}
					isError={completeness.isError}
					onRetry={() => void completeness.refetch()}
				>
					{completeness.data && <CompletenessView data={completeness.data} />}
				</Section>
				<Section
					id="r-checks"
					title="Внешние проверки"
					isLoading={checks.isLoading}
					isError={checks.isError}
					onRetry={() => void checks.refetch()}
					extra={<RefreshChecks applicationId={applicationId} />}
				>
					{checks.data && <ChecksView checks={checks.data} />}
				</Section>
				{(infoRequests.isLoading || infoRequests.isError || (infoRequests.data?.length ?? 0) > 0) && (
					<Section
						id="r-requests"
						title="Запросы информации"
						isLoading={infoRequests.isLoading}
						isError={infoRequests.isError}
						onRetry={() => void infoRequests.refetch()}
					>
						{infoRequests.data && <InfoRequestsView requests={infoRequests.data} />}
					</Section>
				)}
				<Section
					id="r-history"
					title="История"
					isLoading={history.isLoading}
					isError={history.isError}
					onRetry={() => void history.refetch()}
				>
					{history.data && <HistoryView history={history.data} />}
				</Section>
				{data && decidable && (
					<Card className={styles.card} aria-labelledby="r-decision">
						<h2 id="r-decision" className={styles.title}>Решение</h2>
						<div className={styles.decisionActions}>
							<MakeDecision
								applicationId={applicationId}
								requested={{ ...data.request }}
								defaultRate={score.data?.suggested?.rate ?? data.prelim?.prelimRate}
							/>
							<RequestInfo applicationId={applicationId} />
						</div>
					</Card>
				)}
			</div>
			<div className={styles.side}>
				<Section
					id="r-score"
					title="Скоринг"
					isLoading={score.isLoading}
					isError={score.isError}
					onRetry={() => void score.refetch()}
				>
					{score.data && <ScoreView score={score.data} />}
				</Section>
			</div>
		</div>
	);
};
