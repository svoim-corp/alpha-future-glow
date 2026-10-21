import { Page } from '@shared/ui';
import { ApplicationTracker } from '@widgets/application-tracker';
import { useParams } from 'react-router';

export const ApplicationPage = () => {
	const { id = '' } = useParams();

	return (
		<Page title="Статус заявки" subtitle="Всегда видно, что происходит и чей сейчас ход.">
			<ApplicationTracker applicationId={id} />
		</Page>
	);
};
