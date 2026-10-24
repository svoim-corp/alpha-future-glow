import { formatDateTime } from '@shared/lib/format';
import { AsyncBoundary, Icon } from '@shared/ui';
import clsx from 'clsx';

import { useGetTimelineQuery } from '../api/trackingApi';
import { type TimelineStep } from '../model/tracking';
import styles from './TimelineView.module.scss';

/** Итоговые «отрицательные» шаги: отказ и закрытие заявки — красный маркер с крестиком */
const isStop = (step: TimelineStep) => (step.status === 'REJECTED' || step.status === 'CLOSED') && step.state !== 'PENDING';

const DECIDED = new Set(['APPROVED', 'COUNTER_OFFER', 'ISSUED']);

/** Шаг выполнен: пройден или это достигнутый положительный итог («Одобрено», «Иные условия», «Выдано») */
const isDone = (step: TimelineStep) => step.state === 'DONE' || (DECIDED.has(step.status) && step.state === 'CURRENT');

const STATE_TEXT = { DONE: 'выполнено', CURRENT: 'текущий этап', PENDING: 'впереди' } as const;

/** Понятный клиенту статус: что происходит, чей ход, и шаги с датами */
export const TimelineView = ({ applicationId }: { applicationId: string }) => {
	const { data, isLoading, isError, refetch } = useGetTimelineQuery(applicationId);

	return (
		<AsyncBoundary
			isLoading={isLoading}
			isError={isError}
			onRetry={() => void refetch()}
			loadingLabel="Загружаем статус"
			errorText="Не удалось загрузить статус заявки"
		>
			{data && (
				<div>
					<div className={clsx(styles.current, data.current.actor === 'CLIENT' && styles.yourTurn)}>
						<p className={styles.what}>{data.current.whatsHappening}</p>
						<p className={styles.next}>
							{data.current.actor === 'CLIENT' ? 'Ход за вами: ' : 'Ход за банком: '}
							{data.current.nextStep}
						</p>
					</div>
					<ol className={styles.steps} aria-label="Этапы рассмотрения">
						{data.steps.map((step) => (
							<li
								key={step.status}
								className={clsx(
									styles.step,
									isDone(step) && styles.done,
									step.state === 'CURRENT' && !isDone(step) && styles.now,
									isStop(step) && styles.rejected,
								)}
								aria-current={step.state === 'CURRENT' ? 'step' : undefined}
							>
								<span className={styles.marker}>
									{isDone(step) && <Icon name="check" size={14} />}
									{isStop(step) && <Icon name="close" size={14} />}
								</span>
								<span className={styles.label}>
									{step.label}
									<span className="visually-hidden">{` — ${STATE_TEXT[step.state]}`}</span>
								</span>
								{step.at && <span className={styles.date}>{formatDateTime(step.at)}</span>}
							</li>
						))}
					</ol>
				</div>
			)}
		</AsyncBoundary>
	);
};
