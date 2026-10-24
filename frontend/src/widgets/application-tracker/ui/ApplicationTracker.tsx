import {
	OfferView,
	StatusBadge,
	TimelineView,
	useGetApplicationQuery,
	useGetInfoRequestsQuery,
	useGetOfferQuery,
} from '@entities/application';
import { AnswerInfoRequest } from '@features/answer-info-request';
import { DecideOffer } from '@features/decide-offer';
import { skipToken } from '@reduxjs/toolkit/query';
import { baseApi } from '@shared/api';
import { paths } from '@shared/config';
import { formatMoney, formatTerm } from '@shared/lib/format';
import { useAppDispatch } from '@shared/lib/store';
import { AsyncBoundary, ButtonLink, Card, Metrics } from '@shared/ui';
import { useEffect } from 'react';
import { Navigate, useNavigate } from 'react-router';

import styles from './ApplicationTracker.module.scss';

/* Бэк проверяет и оценивает заявку в фоне, поэтому сразу после отправки статус ещё промежуточный */
const IN_PROGRESS = new Set(['SUBMITTED', 'DATA_CHECK', 'SCORING', 'ON_DECISION']);
const POLL_MS = 3000;

const HAS_OFFER = new Set(['APPROVED', 'COUNTER_OFFER', 'REJECTED', 'SIGNING', 'ISSUED']);

export const ApplicationTracker = ({ applicationId }: { applicationId: string }) => {
	const { data: application, isLoading, isError, refetch } = useGetApplicationQuery(applicationId);
	const status = application?.status;
	const dispatch = useAppDispatch();
	const navigate = useNavigate();
	const requests = useGetInfoRequestsQuery(status === 'INFO_REQUESTED' ? applicationId : skipToken);
	const offer = useGetOfferQuery(status && HAS_OFFER.has(status) ? applicationId : skipToken);
	const waiting = status !== undefined && IN_PROGRESS.has(status);

	/* Пока банк работает над заявкой, раз в несколько секунд обновляем её данные вместе с таймлайном и предложением */
	useEffect(() => {
		if (!waiting) return undefined;

		const timer = setInterval(() => {
			dispatch(baseApi.util.invalidateTags([{ type: 'Application', id: applicationId }]));
		}, POLL_MS);

		return () => clearInterval(timer);
	}, [waiting, applicationId, dispatch]);

	const openRequests = (requests.data ?? []).filter((request) => request.status === 'OPEN');

	if (status === 'DRAFT') {
		return <Navigate to={paths.applicationEdit(applicationId)} replace />;
	}

	return (
		<AsyncBoundary
			isLoading={isLoading}
			isError={isError}
			onRetry={() => void refetch()}
			loadingLabel="Загружаем заявку"
			errorText="Не удалось загрузить заявку"
		>
			{application && (
				<div className={styles.tracker}>
					<Card className={styles.card} aria-labelledby="t-summary">
						<div className={styles.head}>
							<div>
								<span className={styles.eyebrow}>ЗАЯВКА НА КРЕДИТ</span>
								<h2 id="t-summary" className={styles.title}>
									{application.number ? `Заявка ${application.number}` : 'Заявка'}
								</h2>
								{application.company && <p className={styles.company}>{application.company.name}</p>}
							</div>
							<StatusBadge status={application.status} />
						</div>
						<Metrics
							label="Параметры заявки"
							items={[
								{ label: 'Продукт', value: application.product.name },
								{ label: 'Сумма', value: formatMoney(application.params.amount) },
								...(application.params.term === null
									? []
									: [{ label: 'Срок', value: formatTerm(application.params.term) }]),
							]}
						/>
					</Card>
					<Card className={styles.card} aria-labelledby="t-status">
						<h2 id="t-status" className={styles.title}>Что происходит с заявкой</h2>
						<TimelineView applicationId={application.id} />
					</Card>
					{openRequests.map((request) => (
						<Card key={request.id} className={styles.card} aria-labelledby={`t-req-${request.id}`}>
							<h2 id={`t-req-${request.id}`} className={styles.title}>Банк просит уточнить</h2>
							<AnswerInfoRequest applicationId={application.id} request={request} />
						</Card>
					))}
					{offer.data && (
						<Card className={styles.card} aria-labelledby="t-offer">
							<h2 id="t-offer" className={styles.title}>Решение банка</h2>
							<OfferView offer={offer.data} />
							{(application.status === 'APPROVED' || application.status === 'COUNTER_OFFER') && (
								<DecideOffer
									applicationId={application.id}
									onAccepted={(redirect) => void navigate(redirect)}
								/>
							)}
							{application.status === 'SIGNING' && (
								<div>
									<ButtonLink to={paths.applicationSign(application.id)}>
										Перейти к подписанию
									</ButtonLink>
								</div>
							)}
						</Card>
					)}
					<div>
						<ButtonLink variant="ghost" to={paths.support(application.id)}>
							Задать вопрос по заявке
						</ButtonLink>
					</div>
				</div>
			)}
		</AsyncBoundary>
	);
};
