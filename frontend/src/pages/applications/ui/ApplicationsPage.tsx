import { ROUTES } from '@shared/config';
import { ButtonLink, Icon, Page } from '@shared/ui';
import { MyApplications } from '@widgets/my-applications';

export const ApplicationsPage = () => {
	return (
		<Page
			title="Мои заявки"
			subtitle="Каждое обращение — с понятным статусом и следующим шагом."
			actions={<ButtonLink to={ROUTES.products}>Новая заявка <Icon name="arrow" size={17} /></ButtonLink>}
		>
			<MyApplications />
		</Page>
	);
};
