import { UploadDocument } from '@features/manage-documents';
import { Page } from '@shared/ui';
import { DocumentsLibrary } from '@widgets/documents-library';

export const DocumentsPage = () => (
	<Page
		title="Документы"
		subtitle="Храните один раз. Используйте в новых заявках."
		actions={<UploadDocument />}
	>
		<DocumentsLibrary />
	</Page>
);
