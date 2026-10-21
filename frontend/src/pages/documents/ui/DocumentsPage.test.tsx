import { documentsFixture } from '@test/fixtures/documents';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { DocumentsPage } from './DocumentsPage';

describe('DocumentsPage', () => {
	it('показывает файлы и счётчики категорий', async () => {
		server.use(http.get(apiUrl('/documents'), () => HttpResponse.json(documentsFixture)));
		renderWithProviders(<DocumentsPage />);

		expect(await screen.findByRole('article', { name: 'Реквизиты компании' })).toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'Все файлы 6' })).toHaveAttribute('aria-pressed', 'true');
		expect(screen.getByRole('button', { name: 'Добавить документ' })).toBeInTheDocument();
	});

	it('категория и поиск уходят в запрос', async () => {
		const query = vi.fn();
		server.use(http.get(apiUrl('/documents'), ({ request }) => {
			query(Object.fromEntries(new URL(request.url).searchParams));

			return HttpResponse.json(documentsFixture);
		}));
		const { user } = renderWithProviders(<DocumentsPage />);

		await screen.findByRole('article', { name: 'Реквизиты компании' });
		await user.click(screen.getByRole('button', { name: /^Фото/ }));
		await user.type(screen.getByLabelText('Поиск документов'), 'склад');

		await waitFor(() => expect(query).toHaveBeenLastCalledWith({ category: 'PHOTOS', q: 'склад', sort: 'name' }));
	});

	it('пустой результат поиска отличается от пустой библиотеки', async () => {
		server.use(http.get(apiUrl('/documents'), () => HttpResponse.json({
			items: [],
			counts: { ALL: 0, COMPANY: 0, STATEMENTS: 0, CONTRACTS: 0, PHOTOS: 0 },
		})));
		renderWithProviders(<DocumentsPage />);

		expect(await screen.findByText('Документов пока нет')).toBeInTheDocument();
	});

	it('при ошибке даёт повторить', async () => {
		server.use(
			http.get(apiUrl('/documents'), () => HttpResponse.json({ code: 'X', message: 'x' }, { status: 500 }), { once: true }),
			http.get(apiUrl('/documents'), () => HttpResponse.json(documentsFixture)),
		);
		const { user } = renderWithProviders(<DocumentsPage />);

		await user.click(await screen.findByRole('button', { name: 'Повторить' }));

		expect(await screen.findByRole('article', { name: 'Реквизиты компании' })).toBeInTheDocument();
	});
});
