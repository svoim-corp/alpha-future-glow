import { faqFixture, ticketsFixture } from '@test/fixtures/support';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { chooseOption } from '@test/select';
import { screen, within } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { SupportPage } from './SupportPage';

describe('SupportPage', () => {
	it.each([
		{
			kind: 'faq', endpoint: 'faq', id: 'f-3', tab: 'Частые вопросы',
			title: 'Зачем выписка из 1С?', text: 'Она подтверждает обороты и позволяет получить больший лимит',
			response: faqFixture[0],
		},
		{
			kind: 'ticket', endpoint: 'tickets', id: 't-12', tab: 'Мои обращения',
			title: 'Не загружается выписка', text: 'Выгрузите выписку в формате 1CClientBankExchange',
			response: ticketsFixture[0],
		},
	])('открывает $kind из поиска и сохраняет вкладку после закрытия', async (item) => {
		if (!item.response) throw new Error('Нет тестовых данных поддержки');
		server.use(
			http.get(apiUrl('/support/faq'), () => HttpResponse.json([])),
			http.get(apiUrl('/support/tickets'), () => HttpResponse.json([])),
			http.get(apiUrl(`/support/${item.endpoint}/${item.id}`), () => HttpResponse.json(item.response)),
		);
		const { user } = renderWithProviders(<SupportPage />, { route: `/support?${item.kind}=${item.id}` });
		const dialog = await screen.findByRole('dialog', { name: item.title });
		expect(within(dialog).getByText(item.text)).toBeInTheDocument();
		expect(screen.getByRole('tab', { name: item.tab })).toHaveAttribute('aria-selected', 'true');
		await user.click(within(dialog).getByRole('button', { name: 'Закрыть' }));
		expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
		expect(screen.getByRole('tab', { name: item.tab })).toHaveAttribute('aria-selected', 'true');
	});

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
		await user.click(screen.getByRole('tab', { name: 'Частые вопросы' }));

		expect(await screen.findByText('Зачем выписка из 1С?')).toBeInTheDocument();
		await chooseOption(user, 'Шаг', 'Документы');

		expect(steps).toHaveBeenCalledWith('DOCUMENTS');
	});

	it('номер заявки из ссылки привязывается к обращению', async () => {
		server.use(
			http.get(apiUrl('/support/faq'), () => HttpResponse.json([])),
			http.get(apiUrl('/support/tickets'), () => HttpResponse.json([])),
		);
		renderWithProviders(<SupportPage />, { route: '/support?tab=tickets&applicationId=a-1031' });

		expect(await screen.findByText('Обращение будет привязано к заявке a-1031.')).toBeInTheDocument();
	});

	it('поиск оставляет подходящие вопросы и сообщает, если ответа нет', async () => {
		server.use(
			http.get(apiUrl('/support/faq'), () => HttpResponse.json(faqFixture)),
			http.get(apiUrl('/support/tickets'), () => HttpResponse.json(ticketsFixture)),
		);
		const { user } = renderWithProviders(<SupportPage />);
		await user.click(screen.getByRole('tab', { name: 'Частые вопросы' }));

		await screen.findByText('Зачем выписка из 1С?');
		await user.type(screen.getByRole('searchbox', { name: 'Поиск по вопросам' }), 'такого нет');

		expect(screen.queryByText('Зачем выписка из 1С?')).not.toBeInTheDocument();
		expect(screen.getByText(/Подходящий ответ не найден/)).toBeInTheDocument();
	});
});
