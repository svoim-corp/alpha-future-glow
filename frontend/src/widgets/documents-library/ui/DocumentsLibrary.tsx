import {
	CATEGORIES,
	DocumentCard,
	type DocumentCategory,
	type DocumentSort,
	formatFileSize,
	useGetDocumentsQuery,
} from '@entities/document';
import { AttachDocument, DeleteDocument } from '@features/manage-documents';
import { useDebouncedValue } from '@shared/lib/debounce';
import { AsyncBoundary, Icon, SearchField, SelectFilter } from '@shared/ui';
import clsx from 'clsx';
import { useState } from 'react';

import styles from './DocumentsLibrary.module.scss';

const SORTS = [
	{ value: 'name', label: 'По названию' },
	{ value: '-uploadedAt', label: 'Сначала новые' },
];

const pluralFiles = (count: number) => {
	const rest = count % 100;
	const last = count % 10;

	if (rest > 10 && rest < 20) {
		return `${count} файлов`;
	}

	return last === 1 ? `${count} файл` : last >= 2 && last <= 4 ? `${count} файла` : `${count} файлов`;
};

/** Библиотека по макету documents.html: вводный блок, поиск и сортировка, категории, сетка файлов */
export const DocumentsLibrary = () => {
	const [category, setCategory] = useState<DocumentCategory | 'ALL'>('ALL');
	const [search, setSearch] = useState('');
	const [sort, setSort] = useState<DocumentSort>('name');
	const q = useDebouncedValue(search, 300);
	const { data, isLoading, isError, refetch } = useGetDocumentsQuery({
		category: category === 'ALL' ? undefined : category,
		q,
		sort,
	});
	const filtered = category !== 'ALL' || q.trim() !== '';
	const totalBytes = data?.items.reduce((total, item) => total + item.sizeBytes, 0) ?? 0;

	return (
		<div className={styles.library}>
			<div className={styles.intro}>
				<div className={styles.introMain}>
					<span className={styles.introIcon}><Icon name="documents" size={20} /></span>
					<div>
						<strong>Ваши документы — в одном месте</strong>
						<p>Выбирайте сохранённые файлы при заполнении заявки.</p>
					</div>
				</div>
				{data && (
					<span className={styles.storage}>
						<strong>{pluralFiles(data.counts.ALL)}</strong>
						{!filtered && totalBytes > 0 && ` · ${formatFileSize(totalBytes)}`}
					</span>
				)}
			</div>
			<div className={styles.toolbar}>
				<SearchField
					label="Поиск документов"
					placeholder="Найти документ по названию"
					value={search}
					onChange={setSearch}
				/>
				<SelectFilter
					label="Сортировка"
					value={sort}
					onChange={(event) => setSort(event.target.value as DocumentSort)}
				>
					{SORTS.map((item) => <option key={item.value} value={item.value}>{item.label}</option>)}
				</SelectFilter>
			</div>
			<div className={styles.chips} role="group" aria-label="Категории документов">
				{CATEGORIES.map((item) => (
					<button
						key={item.value}
						type="button"
						className={clsx(styles.chip)}
						aria-pressed={category === item.value}
						onClick={() => setCategory(item.value)}
					>
						{item.label}
						{data && item.value === 'ALL' && <span>{data.counts.ALL}</span>}
					</button>
				))}
			</div>
			<AsyncBoundary
				isLoading={isLoading}
				isError={isError}
				isEmpty={data?.items.length === 0}
				onRetry={() => void refetch()}
				loadingLabel="Загружаем документы"
				errorText="Не удалось загрузить документы"
				emptyTitle={filtered ? 'Файлы не найдены' : 'Документов пока нет'}
				emptyText={filtered
					? 'Измените запрос или выберите другую категорию.'
					: 'Добавьте файл один раз — и используйте его в новых заявках.'}
			>
				<div className={styles.grid}>
					{data?.items.map((document) => (
						<DocumentCard
							key={document.id}
							document={document}
							actions={(
								<>
									<AttachDocument id={document.id} name={document.name} />
									<DeleteDocument id={document.id} name={document.name} />
								</>
							)}
						/>
					))}
				</div>
			</AsyncBoundary>
		</div>
	);
};
