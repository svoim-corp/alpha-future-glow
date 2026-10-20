import { documentsFixture } from '@test/fixtures/documents';
import { renderWithProviders } from '@test/render';
import { screen } from '@testing-library/react';

import { DocumentCard } from './DocumentCard';

describe('DocumentCard', () => {
	it('показывает тип, размер, дату и использование; просмотр — ссылка', () => {
		renderWithProviders(<DocumentCard document={documentsFixture.items[0]!} />);

		expect(screen.getByRole('link', { name: 'Просмотреть: Реквизиты компании' }))
			.toHaveAttribute('href', '/api/v1/documents/doc-1/download');
		expect(screen.getByText('240 КБ · 12.10.2026')).toBeInTheDocument();
		expect(screen.getByText('Использован в 2 заявках')).toBeInTheDocument();
	});

	it('выводит переданные действия', () => {
		renderWithProviders(<DocumentCard document={documentsFixture.items[1]!} actions={<button type="button">Удалить</button>} />);

		expect(screen.getByRole('button', { name: 'Удалить' })).toBeInTheDocument();
		expect(screen.getByText('Ещё не использован')).toBeInTheDocument();
	});
});
