import { ContractTermsView, useGetContractQuery } from '@entities/contract';
import { SignContract } from '@features/sign-contract';
import { AsyncBoundary, Card } from '@shared/ui';

import styles from './ContractSigning.module.scss';

export const ContractSigning = ({ applicationId }: { applicationId: string }) => {
	const { data, isLoading, isError, refetch } = useGetContractQuery(applicationId);

	return (
		<AsyncBoundary
			isLoading={isLoading}
			isError={isError}
			onRetry={() => void refetch()}
			loadingLabel="Загружаем договор"
			errorText="Не удалось загрузить договор"
		>
			{data && (
				<div className={styles.signing}>
					<Card className={styles.card}>
						<h2 className={styles.title}>Условия кредита</h2>
						<ContractTermsView contract={data} />
					</Card>
					<Card className={styles.card}>
						<SignContract applicationId={applicationId} contract={data} />
					</Card>
				</div>
			)}
		</AsyncBoundary>
	);
};
