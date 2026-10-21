import { bankStatsFixture, queuePageFixture } from '@test/fixtures/bank';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { chooseOption } from '@test/select';
import { screen, within } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { BankQueuePage } from './BankQueuePage';

const base = () => server.use(
	http.get(apiUrl('/bank/applications/stats'), () => HttpResponse.json(bankStatsFixture)),
);

describe('BankQueuePage', () => {
	it('показывает счётчики и таблицу со ссылкой на разбор', async () => {
		base();
		server.use(http.get(apiUrl('/bank/applications'), () => HttpResponse.json(queuePageFixture)));
		renderWithProviders(<BankQueuePage />);

		expect(await screen.findByRole('link', { name: 'КЗ-2026-001024' })).toHaveAttribute(
			'href',
			'/bank/applications/a-1024',
		);
		expect(screen.getByText('Решено сегодня')).toBeInTheDocument();
		expect(screen.getByText('Средний риск')).toBeInTheDocument();
	});

	it('фильтр по статусу уходит в запрос и сбрасывает страницу', async () => {
		base();
		const query = vi.fn();
		server.use(http.get(apiUrl('/bank/applications'), ({ request }) => {
			query(Object.fromEntries(new URL(request.url).searchParams));

			return HttpResponse.json(queuePageFixture);
		}));
		const { user } = renderWithProviders(<BankQueuePage />);

		await screen.findByRole('link', { name: 'КЗ-2026-001024' });
		await chooseOption(user, 'Статус', 'На решении');

		await screen.findByRole('link', { name: 'КЗ-2026-001024' });
		expect(query).toHaveBeenLastCalledWith(expect.objectContaining({ status: 'ON_DECISION', page: '1' }));
	});

	it('пустая выдача объясняет, что менять', async () => {
		base();
		server.use(http.get(
			apiUrl('/bank/applications'),
			() => HttpResponse.json({ items: [], page: 1, size: 20, total: 0 }),
		));
		renderWithProviders(<BankQueuePage />);

		expect(await screen.findByText('По этим условиям заявок нет')).toBeInTheDocument();
	});

	it('пагинация: на первой странице «Назад» недоступна, «Вперёд» ведёт дальше', async () => {
		base();
		const pages = vi.fn();
		server.use(http.get(apiUrl('/bank/applications'), ({ request }) => {
			const page = new URL(request.url).searchParams.get('page');
			pages(page);

			return HttpResponse.json({ ...queuePageFixture, page: Number(page), total: 45 });
		}));
		const { user } = renderWithProviders(<BankQueuePage />);

		const pager = await screen.findByRole('navigation', { name: 'Страницы очереди' });
		expect(within(pager).getByRole('button', { name: 'Назад' })).toBeDisabled();
		await user.click(within(pager).getByRole('button', { name: 'Вперёд' }));

		await screen.findByText('Страница 2 из 3');
		expect(pages).toHaveBeenLastCalledWith('2');
	});

	it('ошибка таблицы не скрывает счётчики и даёт повторить', async () => {
		base();
		server.use(http.get(apiUrl('/bank/applications'), () => HttpResponse.json(
			{ code: 'INTERNAL', message: 'x' },
			{ status: 500 },
		)));
		renderWithProviders(<BankQueuePage />);

		expect(await screen.findByText('Не удалось загрузить очередь')).toBeInTheDocument();
		expect(await screen.findByText('Решено сегодня')).toBeInTheDocument();
	});
});
