import { useGetMeQuery } from '@entities/session';
import { ROUTES } from '@shared/config';
import { ButtonLink, Icon, Page } from '@shared/ui';
import { MyLoans } from '@widgets/my-loans';

export const CreditsPage = () => {
	const { data: user } = useGetMeQuery();
	const company = user?.company?.name.match(/«(.+)»/)?.[1] ?? user?.company?.name;

	return (
		<Page
			eyebrow={`КРЕДИТНЫЙ КАБИНЕТ${company ? ` · ${company.toUpperCase()}` : ''}`}
			title="Мои кредиты"
			subtitle="Обязательства под контролем, планы — в движении."
			actions={<ButtonLink to={ROUTES.products}>Подобрать кредит <Icon name="arrow" size={17} /></ButtonLink>}
		>
			<MyLoans />
		</Page>
	);
};
