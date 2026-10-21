export type DocumentCategory = 'COMPANY' | 'STATEMENTS' | 'CONTRACTS' | 'PHOTOS';
export type DocumentExtension = 'PDF' | 'TXT' | 'JPG' | 'PNG';
export type DocumentSort = 'name' | '-uploadedAt';

export interface LibraryDocument {
	id: string;
	name: string;
	category: DocumentCategory;
	extension: DocumentExtension;
	sizeBytes: number;
	uploadedAt: string;
	usage: { applications: number; label: string | null };
}

export interface DocumentCounts {
	ALL: number;
	COMPANY: number;
	STATEMENTS: number;
	CONTRACTS: number;
	PHOTOS: number;
}

export interface DocumentsResponse {
	items: LibraryDocument[];
	counts: DocumentCounts;
}

export interface DocumentFilters {
	category?: DocumentCategory;
	q?: string;
	sort: DocumentSort;
}

export const CATEGORIES: { value: DocumentCategory | 'ALL'; label: string }[] = [
	{ value: 'ALL', label: 'Все файлы' },
	{ value: 'COMPANY', label: 'Компания' },
	{ value: 'STATEMENTS', label: 'Выписки' },
	{ value: 'CONTRACTS', label: 'Договоры' },
	{ value: 'PHOTOS', label: 'Фото' },
];
