import { draftApplication, infoRequestedApplication } from '@test/fixtures/applications';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { chooseOption } from '@test/select';
import { screen, within } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { MyApplications } from './MyApplications';

const list = (items: unknown[]) => HttpResponse.json({ items, page: 1, size: 20, total: items.length });

const approved = {
	...infoRequestedApplication,
	id: 'a-1021',
	number: 'КЗ-2026-001021',
	product: { code: 'EXPRESS', name: 'Экспресс-кредит' },
	amount: 800000,
	term: 12,
	status: 'APPROVED',
	statusLabel: 'Одобрено',
	updatedAt: '2026-10-18T09:00:00Z',
};

const setup = () => {
	server.use(http.get(apiUrl('/applications'), () => list([infoRequestedApplication, approved, draftApplication])));

	return renderWithProviders(<MyApplications />);
};

describe('MyApplications', () => {
	it('считает обзор: заявки без черновиков, требующие решения, одобренные', async () => {
		setup();

		const overview = await screen.findByRole('group', { name: 'Обзор заявок' }).catch(() => null);
		const totals = overview ?? (await screen.findByText('Всего заявок')).closest('dl')!;

		expect(within(totals).getByText('Без учёта черновиков')).toBeInTheDocument();
		expect(screen.getByRole('tab', { name: /Заявки\s*2/ })).toHaveAttribute('aria-selected', 'true');
		expect(screen.getByRole('tab', { name: /Черновики\s*1/ })).toBeInTheDocument();
	});

	it('строка заявки: номер, дата, сумма, статус и следующее действие', async () => {
		setup();

		const rows = await screen.findAllByRole('listitem');
		const first = within(rows[0]!);

		expect(first.getByText('Оборотный кредит')).toBeInTheDocument();
		expect(first.getByText('Нужна информация')).toBeInTheDocument();
		expect(first.getByRole('link', { name: /Дополнить/ })).toHaveAttribute('href', '/applications/a-1024');
		expect(within(rows[1]!).getByRole('link', { name: /Посмотреть/ })).toBeInTheDocument();
	});

	it('поиск, фильтр по статусу и сброс', async () => {
		const { user } = setup();

		await screen.findAllByRole('listitem');
		await user.type(screen.getByRole('searchbox', { name: 'Поиск по заявкам' }), 'экспресс');

		expect(screen.getAllByRole('listitem')).toHaveLength(1);
		expect(screen.getByRole('status')).toHaveTextContent('Найдено: 1 из 2');

		await user.clear(screen.getByRole('searchbox'));
		await chooseOption(user, 'Статус', 'Нужна информация');

		expect(screen.getAllByRole('listitem')).toHaveLength(1);

		await user.type(screen.getByRole('searchbox'), 'ничего');

		expect(screen.getByText('Таких заявок пока нет')).toBeInTheDocument();

		await user.click(screen.getByRole('button', { name: 'Сбросить фильтры' }));

		expect(screen.getAllByRole('listitem')).toHaveLength(2);
	});

	it('сортировка по сумме', async () => {
		const { user } = setup();

		await screen.findAllByRole('listitem');
		await chooseOption(user, 'Сортировка заявок', 'По сумме: меньше');

		expect(within(screen.getAllByRole('listitem')[0]!).getByText('Экспресс-кредит')).toBeInTheDocument();
	});

	it('вкладка «Черновики»: продолжить и удалить', async () => {
		const { user } = setup();

		await user.click(await screen.findByRole('tab', { name: /Черновики/ }));

		const drafts = screen.getByRole('list', { name: 'Черновики' });

		expect(within(drafts).getByText('Экспресс-кредит')).toBeInTheDocument();
		expect(within(drafts).getByRole('link', { name: /Продолжить заполнение/ })).toHaveAttribute(
			'href',
			'/applications/a-1031/edit',
		);
		expect(within(drafts).getByRole('button', { name: 'Удалить черновик' })).toBeInTheDocument();
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
			http.get(apiUrl('/applications'), () => list([approved])),
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
