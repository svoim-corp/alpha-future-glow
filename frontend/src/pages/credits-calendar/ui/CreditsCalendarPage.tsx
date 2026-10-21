import { ROUTES } from '@shared/config';
import { ButtonLink, Page } from '@shared/ui';
import { PaymentCalendar } from '@widgets/payment-calendar';

export const CreditsCalendarPage = () => (
	<Page
		title="Платёжный календарь"
		subtitle="Все платежи по кредитам на одном экране."
		actions={<ButtonLink variant="secondary" to={ROUTES.credits}>Мои кредиты</ButtonLink>}
	>
		<PaymentCalendar />
	</Page>
);
