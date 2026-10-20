import { LoansSummaryCard } from '@entities/loan';
import { ROUTES } from '@shared/config';
import { ButtonLink, Page } from '@shared/ui';
import { MyApplications } from '@widgets/my-applications';

import styles from './HomePage.module.scss';

export const HomePage = () => (
	<Page
		title="Мои заявки"
		subtitle="Статусы заявок и черновики. Всё, что вы ввели, сохраняется."
		actions={<ButtonLink to={ROUTES.products}>Получить финансирование</ButtonLink>}
	>
		<div className={styles.stack}>
			<LoansSummaryCard />
			<MyApplications />
		</div>
	</Page>
);
