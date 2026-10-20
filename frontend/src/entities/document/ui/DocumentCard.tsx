import { formatDate } from '@shared/lib/format';
import { Card } from '@shared/ui';
import { type ReactNode } from 'react';

import { formatFileSize, usageLabel } from '../model/format';
import { type LibraryDocument } from '../model/types';
import styles from './DocumentCard.module.scss';

interface DocumentCardProps {
	document: LibraryDocument;
	/** Действия над файлом: «В заявку», «Удалить» — собираются выше, фичи не импортируют друг друга */
	actions?: ReactNode;
}

export const DocumentCard = ({ document, actions }: DocumentCardProps) => (
	<Card as="article" className={styles.card} aria-label={document.name}>
		<a
			className={styles.open}
			href={`/api/v1/documents/${document.id}/download`}
			target="_blank"
			rel="noreferrer"
			aria-label={`Просмотреть: ${document.name}`}
		>
			<span className={styles.ext}>{document.extension}</span>
			<span className={styles.name}>{document.name}</span>
		</a>
		<p className={styles.meta}>{`${formatFileSize(document.sizeBytes)} · ${formatDate(document.uploadedAt)}`}</p>
		<p className={styles.meta}>{usageLabel(document.usage)}</p>
		{actions && <div className={styles.actions}>{actions}</div>}
	</Card>
);
