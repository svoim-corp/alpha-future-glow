import { ROUTES } from '@shared/config';
import { formatDate, formatMoney } from '@shared/lib/format';
import { AsyncBoundary, ButtonLink, Card } from '@shared/ui';

import { useGetLoansSummaryQuery } from '../api/loanApi';
import styles from './LoansSummaryCard.module.scss';

/** Блок на главной: есть ли выданные кредиты и когда ближайший платёж. Без кредитов не показывается */
export const LoansSummaryCard = () => {
	const { data, isLoading, isError, refetch } = useGetLoansSummaryQuery();

	if (!isLoading && !isError && (!data || data.activeLoans === 0)) {
		return null;
	}

	return (
		<AsyncBoundary isLoading={isLoading} isError={isError} onRetry={() => void refetch()} loadingLabel="Загрузка кредитов">
			{data && (
				<Card aria-labelledby="loans-summary-title" className={styles.card}>
					<div>
						<h2 id="loans-summary-title" className={styles.title}>Мои кредиты</h2>
						<p className={styles.text}>
							Действующих кредитов: {data.activeLoans}.
							{data.nextPayment && (
								<>
									{' '}Ближайший платёж {formatDate(data.nextPayment.date)} —{' '}
									<span className={styles.amount}>{formatMoney(data.nextPayment.amount)}</span>
								</>
							)}
						</p>
					</div>
					<ButtonLink to={ROUTES.credits} variant="secondary" compact>Перейти к кредитам</ButtonLink>
				</Card>
			)}
		</AsyncBoundary>
	);
};
