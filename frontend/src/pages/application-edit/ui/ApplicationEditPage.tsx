import { Page } from '@shared/ui';
import { ApplicationWizard } from '@widgets/application-wizard';
import { useParams } from 'react-router';

export const ApplicationEditPage = () => {
	const { id = '' } = useParams();

	return (
		<Page title="Заявка на кредит" subtitle="Изменения сохраняются автоматически — можно вернуться позже.">
			<ApplicationWizard applicationId={id} />
		</Page>
	);
};
