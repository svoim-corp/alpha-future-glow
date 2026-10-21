import { useGetMeQuery } from '@entities/session';
import { ROUTES } from '@shared/config';
import { ButtonLink, Icon, Page } from '@shared/ui';
import { MyApplications } from '@widgets/my-applications';

export const ApplicationsPage = () => {
	const { data: user } = useGetMeQuery();
	const company = user?.company?.name.match(/«(.+)»/)?.[1] ?? user?.company?.name;

	return (
		<Page
			eyebrow={`КРЕДИТНЫЙ КАБИНЕТ${company ? ` · ${company.toUpperCase()}` : ''}`}
			title="Мои заявки"
			subtitle="Каждое обращение — с понятным статусом и следующим шагом."
			actions={<ButtonLink to={ROUTES.products}>Новая заявка <Icon name="arrow" size={17} /></ButtonLink>}
		>
			<MyApplications />
		</Page>
	);
};
