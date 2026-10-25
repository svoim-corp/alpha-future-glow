import { Page } from '@shared/ui';
import { BankQueue } from '@widgets/bank-queue';

export const BankQueuePage = () => (
	<Page
		eyebrow="РАБОЧЕЕ МЕСТО БАНКА"
		title="Очередь заявок"
		subtitle="Заявки, ожидающие проверки и решения."
	>
		<BankQueue />
	</Page>
);
