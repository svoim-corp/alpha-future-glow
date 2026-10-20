import { ContractTermsView, useGetContractQuery } from '@entities/contract';
import { SignContract } from '@features/sign-contract';
import { AsyncBoundary, Card } from '@shared/ui';

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
				<>
					<Card>
						<h2>Условия кредита</h2>
						<ContractTermsView contract={data} />
					</Card>
					<Card>
						<SignContract applicationId={applicationId} contract={data} />
					</Card>
				</>
			)}
		</AsyncBoundary>
	);
};
