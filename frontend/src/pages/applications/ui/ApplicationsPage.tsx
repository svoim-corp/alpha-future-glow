import { Page } from '@shared/ui';
import { MyApplications } from '@widgets/my-applications';

export const ApplicationsPage = () => (
	<Page title="Мои заявки" subtitle="Понятный путь от заявки до решения.">
		<MyApplications />
	</Page>
);
