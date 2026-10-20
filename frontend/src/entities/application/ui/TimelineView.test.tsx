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

	it('при ошибке показывает сообщение', async () => {
		server.use(http.get(apiUrl('/applications/a-1/timeline'), () => HttpResponse.json(
			{ code: 'INTERNAL', message: 'Ошибка' },
			{ status: 500 },
		)));
		renderWithProviders(<TimelineView applicationId="a-1" />);

		expect(await screen.findByText('Не удалось загрузить статус заявки')).toBeInTheDocument();
	});
});
