import { formatDate } from '@shared/lib/format';
import clsx from 'clsx';
import { type ReactNode } from 'react';

import { formatFileSize, usageLabel } from '../model/format';
import { type DocumentExtension, type LibraryDocument } from '../model/types';
import styles from './DocumentCard.module.scss';

interface DocumentCardProps {
	document: LibraryDocument;
	/** Действия над файлом: «В заявку», «Удалить» — собираются выше, фичи не импортируют друг друга */
	actions?: ReactNode;
}

const isPhoto = (extension: DocumentExtension) => extension === 'JPG' || extension === 'PNG';

/** Карточка файла по макету documents.html: обложка с превью, название, размер и дата, использование, действия */
export const DocumentCard = ({ document, actions }: DocumentCardProps) => (
	<article className={styles.card} aria-label={document.name}>
		<a
			className={styles.open}
			href={`/api/v1/documents/${document.id}/download`}
			target="_blank"
			rel="noreferrer"
			aria-label={`Просмотреть: ${document.name}`}
		>
			<div className={clsx(styles.cover, document.extension === 'TXT' && styles.txt, isPhoto(document.extension) && styles.photo)}>
				{isPhoto(document.extension) ? (
					<div className={styles.placeholder} aria-hidden="true"><span /><i /><i /><i /></div>
				) : (
					<div className={styles.paper} aria-hidden="true">
						<span>{document.category === 'CONTRACTS' ? 'КРЕДИТНЫЙ ДОГОВОР' : 'ДОКУМЕНТ КОМПАНИИ'}</span>
						<i /><i /><i /><i />
					</div>
				)}
				<span className={styles.ext}>{document.extension}</span>
			</div>
			<span className={styles.name}>{document.name}</span>
		</a>
		<div className={styles.meta}>
			<span>{`${formatFileSize(document.sizeBytes)} · ${formatDate(document.uploadedAt)}`}</span>
			<small>{usageLabel(document.usage)}</small>
		</div>
		{actions && <div className={styles.actions}>{actions}</div>}
	</article>
);
