import { Page } from '@shared/ui';
import { ContractSigning } from '@widgets/contract-signing';
import { useParams } from 'react-router';

export const ApplicationSignPage = () => {
	const { id = '' } = useParams();

	return (
		<Page title="Подписание договора" subtitle="Проверьте условия и документы. Подпись имитируется, без кода из SMS.">
			<ContractSigning applicationId={id} />
		</Page>
	);
};
