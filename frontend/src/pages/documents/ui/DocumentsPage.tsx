import { useGetMeQuery } from '@entities/session';
import { UploadDocument } from '@features/manage-documents';
import { Page } from '@shared/ui';
import { DocumentsLibrary } from '@widgets/documents-library';

export const DocumentsPage = () => {
	const { data: user } = useGetMeQuery();
	const company = user?.company?.name.match(/«(.+)»/)?.[1] ?? user?.company?.name;

	return (
		<Page
			eyebrow={`КРЕДИТНЫЙ КАБИНЕТ${company ? ` · ${company.toUpperCase()}` : ''}`}
			title="Документы"
			subtitle="Храните один раз. Используйте в новых заявках."
			actions={<UploadDocument />}
		>
			<DocumentsLibrary />
		</Page>
	);
};
