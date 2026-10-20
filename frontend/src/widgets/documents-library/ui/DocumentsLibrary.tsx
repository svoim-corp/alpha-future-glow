import {
	CATEGORIES,
	DocumentCard,
	type DocumentCategory,
	type DocumentSort,
	useGetDocumentsQuery,
} from '@entities/document';
import { AttachDocument, DeleteDocument } from '@features/manage-documents';
import { useDebouncedValue } from '@shared/lib/debounce';
import { AsyncBoundary, Field, SelectField } from '@shared/ui';
import { useState } from 'react';

import styles from './DocumentsLibrary.module.scss';

const SORTS = [
	{ value: 'name', label: 'По названию' },
	{ value: '-uploadedAt', label: 'Сначала новые' },
];

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

	return (
		<div className={styles.library}>
			<div className={styles.toolbar}>
				<Field
					label="Поиск документов"
					type="search"
					placeholder="Найти документ по названию"
					value={search}
					onChange={(event) => setSearch(event.target.value)}
				/>
				<SelectField
					label="Сортировка"
					value={sort}
					options={SORTS}
					onChange={(event) => setSort(event.target.value as DocumentSort)}
				/>
			</div>
			<div className={styles.chips} role="group" aria-label="Категории документов">
				{CATEGORIES.map((item) => (
					<button
						key={item.value}
						type="button"
						className={styles.chip}
						aria-pressed={category === item.value}
						onClick={() => setCategory(item.value)}
					>
						{data ? `${item.label} ${data.counts[item.value === 'ALL' ? 'ALL' : item.value]}` : item.label}
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
				emptyTitle={filtered ? 'Ничего не найдено' : 'Документов пока нет'}
				emptyText={filtered
					? 'Измените поиск или выберите другую категорию.'
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
