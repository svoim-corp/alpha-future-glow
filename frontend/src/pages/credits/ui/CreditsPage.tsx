import { ROUTES } from '@shared/config';
import { ButtonLink, Page } from '@shared/ui';
import { MyLoans } from '@widgets/my-loans';

export const CreditsPage = () => (
	<Page
		title="Мои кредиты"
		subtitle="Остаток долга, условия договора и график платежей."
		actions={<ButtonLink variant="secondary" to={ROUTES.creditsCalendar}>Платёжный календарь</ButtonLink>}
	>
		<MyLoans />
	</Page>
);
