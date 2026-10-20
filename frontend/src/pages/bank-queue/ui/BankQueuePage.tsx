import { Page } from '@shared/ui';
import { BankQueue } from '@widgets/bank-queue';

export const BankQueuePage = () => (
	<Page title="Очередь заявок" subtitle="Заявки, ожидающие проверки и решения.">
		<BankQueue />
	</Page>
);
