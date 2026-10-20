import { draftApplication, infoRequestedApplication } from '@test/fixtures/applications';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen, within } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { MyApplications } from './MyApplications';

const list = (items: unknown[]) => HttpResponse.json({ items, page: 1, size: 20, total: items.length });

describe('MyApplications', () => {
	it('показывает заявки: у черновика — удаление, у поданной — бейдж действия', async () => {
		server.use(http.get(apiUrl('/applications'), () => list([infoRequestedApplication, draftApplication])));
		renderWithProviders(<MyApplications />);

		const items = await screen.findAllByRole('listitem');

		expect(items).toHaveLength(2);
		expect(within(items[0]!).getByText('Нужно действие')).toBeInTheDocument();
		expect(within(items[0]!).queryByRole('button', { name: 'Удалить черновик' })).not.toBeInTheDocument();
		expect(within(items[1]!).getByRole('button', { name: 'Удалить черновик' })).toBeInTheDocument();
	});

	it('без заявок предлагает начать с каталога', async () => {
		server.use(http.get(apiUrl('/applications'), () => list([])));
		renderWithProviders(<MyApplications />);

		expect(await screen.findByText('Заявок пока нет')).toBeInTheDocument();
		expect(screen.getByRole('link', { name: 'Получить финансирование' })).toHaveAttribute('href', '/products');
	});

	it('при ошибке показывает алерт и умеет перезапросить', async () => {
		server.use(
			http.get(apiUrl('/applications'), () => HttpResponse.json(
				{ code: 'UPSTREAM_UNAVAILABLE', message: 'Мок недоступен' },
				{ status: 502 },
			), { once: true }),
			http.get(apiUrl('/applications'), () => list([draftApplication])),
		);
		const { user } = renderWithProviders(<MyApplications />);

		expect(await screen.findByRole('alert')).toHaveTextContent('Не удалось загрузить заявки.');
		await user.click(screen.getByRole('button', { name: 'Повторить' }));

		expect(await screen.findAllByRole('listitem')).toHaveLength(1);
	});

	it('показывает загрузку', () => {
		server.use(http.get(apiUrl('/applications'), () => list([])));
		renderWithProviders(<MyApplications />);

		expect(screen.getByRole('status', { name: 'Загрузка заявок' })).toBeInTheDocument();
	});
});
