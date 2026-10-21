import { type DocumentsResponse } from '@entities/document';

export const documentsFixture: DocumentsResponse = {
	items: [
		{
			id: 'doc-1',
			name: 'Реквизиты компании',
			category: 'COMPANY',
			extension: 'PDF',
			sizeBytes: 245760,
			uploadedAt: '2026-10-12T09:00:00Z',
			usage: { applications: 2, label: null },
		},
		{
			id: 'doc-5',
			name: 'Фото склада',
			category: 'PHOTOS',
			extension: 'JPG',
			sizeBytes: 2516582,
			uploadedAt: '2026-10-18T09:00:00Z',
			usage: { applications: 0, label: null },
		},
	],
	counts: { ALL: 6, COMPANY: 2, STATEMENTS: 1, CONTRACTS: 2, PHOTOS: 1 },
};
