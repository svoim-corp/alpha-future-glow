import { faqFixture, ticketsFixture } from '@test/fixtures/support';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { SupportPage } from './SupportPage';

describe('SupportPage', () => {
	it('меняет вопросы по выбранному шагу', async () => {
		const steps = vi.fn();
		server.use(
			http.get(apiUrl('/support/faq'), ({ request }) => {
				steps(new URL(request.url).searchParams.get('step'));

				return HttpResponse.json(faqFixture);
			}),
			http.get(apiUrl('/support/tickets'), () => HttpResponse.json(ticketsFixture)),
		);
		const { user } = renderWithProviders(<SupportPage />);

		expect(await screen.findByText('Зачем выписка из 1С?')).toBeInTheDocument();
		await user.selectOptions(screen.getByLabelText('Шаг'), 'DOCUMENTS');

		expect(steps).toHaveBeenCalledWith('DOCUMENTS');
	});

	it('номер заявки из ссылки привязывается к обращению', async () => {
		server.use(
			http.get(apiUrl('/support/faq'), () => HttpResponse.json([])),
			http.get(apiUrl('/support/tickets'), () => HttpResponse.json([])),
		);
		renderWithProviders(<SupportPage />, { route: '/support?applicationId=a-1031' });

		expect(await screen.findByText('Обращение будет привязано к заявке a-1031.')).toBeInTheDocument();
	});

	it('поиск оставляет подходящие вопросы и сообщает, если ответа нет', async () => {
		server.use(
			http.get(apiUrl('/support/faq'), () => HttpResponse.json(faqFixture)),
			http.get(apiUrl('/support/tickets'), () => HttpResponse.json(ticketsFixture)),
		);
		const { user } = renderWithProviders(<SupportPage />);

		await screen.findByText('Зачем выписка из 1С?');
		await user.type(screen.getByRole('searchbox', { name: 'Поиск по вопросам' }), 'такого нет');

		expect(screen.queryByText('Зачем выписка из 1С?')).not.toBeInTheDocument();
		expect(screen.getByText(/Подходящий ответ не найден/)).toBeInTheDocument();
	});
});
