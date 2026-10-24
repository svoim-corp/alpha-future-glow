import { timelineFixture } from '@test/fixtures/tracking';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen, within } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { TimelineView } from './TimelineView';

describe('TimelineView', () => {
	it('объясняет, что происходит и чей ход', async () => {
		server.use(http.get(apiUrl('/applications/a-1/timeline'), () => HttpResponse.json(timelineFixture)));
		renderWithProviders(<TimelineView applicationId="a-1" />);

		expect(await screen.findByText('Банк проверил заявку и просит уточнить данные')).toBeInTheDocument();
		expect(screen.getByText('Ход за вами: Ответьте на запрос банка ниже')).toBeInTheDocument();
	});

	it('отмечает текущий этап не только цветом', async () => {
		server.use(http.get(apiUrl('/applications/a-1/timeline'), () => HttpResponse.json(timelineFixture)));
		renderWithProviders(<TimelineView applicationId="a-1" />);

		const steps = within(await screen.findByRole('list', { name: 'Этапы рассмотрения' })).getAllByRole('listitem');

		expect(steps).toHaveLength(4);
		expect(steps[2]).toHaveAttribute('aria-current', 'step');
		expect(steps[2]).toHaveTextContent('текущий этап');
		expect(steps[3]).toHaveTextContent('впереди');
	});

	it('итог «Отказ» помечен крестиком в маркере', async () => {
		const rejected = {
			...timelineFixture,
			steps: [
				{ status: 'ON_DECISION', label: 'На решении', state: 'DONE', at: '2026-10-20T10:00:00Z' },
				{ status: 'REJECTED', label: 'Отказ', state: 'CURRENT', at: '2026-10-20T11:00:00Z' },
			],
		};
		server.use(http.get(apiUrl('/applications/a-1/timeline'), () => HttpResponse.json(rejected)));
		renderWithProviders(<TimelineView applicationId="a-1" />);

		const steps = within(await screen.findByRole('list', { name: 'Этапы рассмотрения' })).getAllByRole('listitem');

		expect(steps).toHaveLength(2);
		expect(steps[1]?.querySelector('svg path[d="m6 6 12 12M6 18 18 6"]')).not.toBeNull();
		expect(steps[0]?.querySelector('svg path[d="m6 6 12 12M6 18 18 6"]')).toBeNull();
	});

	it('итог «Закрыта» помечен крестиком в маркере', async () => {
		const closed = {
			...timelineFixture,
			steps: [
				{ status: 'APPROVED', label: 'Одобрено', state: 'DONE', at: '2026-10-20T10:00:00Z' },
				{ status: 'CLOSED', label: 'Закрыта', state: 'CURRENT', at: '2026-10-20T11:00:00Z' },
			],
		};
		server.use(http.get(apiUrl('/applications/a-1/timeline'), () => HttpResponse.json(closed)));
		renderWithProviders(<TimelineView applicationId="a-1" />);

		const steps = within(await screen.findByRole('list', { name: 'Этапы рассмотрения' })).getAllByRole('listitem');

		expect(steps).toHaveLength(2);
		expect(steps[1]?.querySelector('svg path[d="m6 6 12 12M6 18 18 6"]')).not.toBeNull();
		expect(steps[0]?.querySelector('svg path[d="m6 6 12 12M6 18 18 6"]')).toBeNull();
	});

	it.each([['ISSUED', 'Выдано'], ['APPROVED', 'Одобрено']])('итог «%s» помечен галочкой', async (status, label) => {
		const issued = {
			...timelineFixture,
			steps: [
				{ status: 'SIGNING', label: 'Подписание', state: 'DONE', at: '2026-10-20T10:00:00Z' },
				{ status, label, state: 'CURRENT', at: '2026-10-20T11:00:00Z' },
			],
		};
		server.use(http.get(apiUrl('/applications/a-1/timeline'), () => HttpResponse.json(issued)));
		renderWithProviders(<TimelineView applicationId="a-1" />);

		const steps = within(await screen.findByRole('list', { name: 'Этапы рассмотрения' })).getAllByRole('listitem');

		expect(steps).toHaveLength(2);
		expect(steps[1]?.querySelector('svg')).not.toBeNull();
		expect(steps[1]?.querySelector('svg path[d="m6 6 12 12M6 18 18 6"]')).toBeNull();
	});

	it('при ошибке показывает сообщение', async () => {
		server.use(http.get(apiUrl('/applications/a-1/timeline'), () => HttpResponse.json(
			{ code: 'INTERNAL', message: 'Ошибка' },
			{ status: 500 },
		)));
		renderWithProviders(<TimelineView applicationId="a-1" />);

		expect(await screen.findByText('Не удалось загрузить статус заявки')).toBeInTheDocument();
	});
});
