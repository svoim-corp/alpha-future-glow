import { faqFixture, ticketsFixture } from '@test/fixtures/support';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { FaqList } from './FaqList';
import { TicketList } from './TicketList';

describe('FaqList', () => {
	it('показывает вопросы выбранного шага', async () => {
		const step = vi.fn();
		server.use(http.get(apiUrl('/support/faq'), ({ request }) => {
			step(new URL(request.url).searchParams.get('step'));

			return HttpResponse.json(faqFixture);
		}));
		renderWithProviders(<FaqList step="APPLICATION" />);

		expect(await screen.findByText('Зачем выписка из 1С?')).toBeInTheDocument();
		expect(step).toHaveBeenCalledWith('APPLICATION');
	});
});

describe('TicketList', () => {
	it('показывает обращения со статусом и ответом', async () => {
		server.use(http.get(apiUrl('/support/tickets'), () => HttpResponse.json(ticketsFixture)));
		renderWithProviders(<TicketList />);

		expect(await screen.findByText('Не загружается выписка')).toBeInTheDocument();
		expect(screen.getByText('Есть ответ')).toBeInTheDocument();
		expect(screen.getByText(/Выгрузите выписку/)).toBeInTheDocument();
	});

	it('без обращений показывает пустое состояние', async () => {
		server.use(http.get(apiUrl('/support/tickets'), () => HttpResponse.json([])));
		renderWithProviders(<TicketList />);

		expect(await screen.findByText('Обращений пока нет')).toBeInTheDocument();
	});
});
