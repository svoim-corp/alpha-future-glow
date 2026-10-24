import { ROUTES } from '@shared/config';
import { ButtonLink, Icon, Page } from '@shared/ui';
import { MyLoans } from '@widgets/my-loans';

export const CreditsPage = () => {
	return (
		<Page
			title="Мои кредиты"
			subtitle="Обязательства под контролем, планы — в движении."
			actions={<ButtonLink to={ROUTES.products}>Подобрать кредит <Icon name="arrow" size={17} /></ButtonLink>}
		>
			<MyLoans />
		</Page>
	);
};
