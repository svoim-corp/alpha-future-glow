import { ROUTES } from '@shared/config';
import { ButtonLink, Page } from '@shared/ui';
import { BankReview } from '@widgets/bank-review';
import { useParams } from 'react-router';

export const BankApplicationPage = () => {
	const { id = '' } = useParams();

	return (
		<Page
			title="Разбор заявки"
			subtitle="Анкета, проверки, скоринг и решение."
			actions={<ButtonLink variant="secondary" to={ROUTES.bankQueue}>К очереди</ButtonLink>}
		>
			<BankReview applicationId={id} />
		</Page>
	);
};
