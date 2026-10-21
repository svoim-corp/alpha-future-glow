import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { CreateTicket } from './CreateTicket';

describe('CreateTicket', () => {
	it('проверяет поля до отправки', async () => {
		const { user } = renderWithProviders(<CreateTicket />);

		await user.click(screen.getByRole('button', { name: 'Отправить обращение' }));

		expect(await screen.findByText('Опишите тему в нескольких словах')).toBeInTheDocument();
		expect(screen.getByText(/хотя бы в одном предложении/)).toBeInTheDocument();
	});

	it('отправляет обращение с номером заявки и очищает форму', async () => {
		const body = vi.fn();
		server.use(http.post(apiUrl('/support/tickets'), async ({ request }) => {
			body(await request.json());

			return HttpResponse.json({ id: 't-1', status: 'OPEN', createdAt: '2026-10-20T11:00:00Z' }, { status: 201 });
		}));
		const { user } = renderWithProviders(<CreateTicket applicationId="a-1031" />);

		await user.type(screen.getByLabelText('Тема'), 'Не загружается выписка');
		await user.type(screen.getByLabelText('Сообщение'), 'Файл из 1С выдаёт ошибку формата');
		await user.click(screen.getByRole('button', { name: 'Отправить обращение' }));

		expect(await screen.findByText('Обращение отправлено')).toBeInTheDocument();
		expect(body).toHaveBeenCalledWith({
			subject: 'Не загружается выписка',
			text: 'Файл из 1С выдаёт ошибку формата',
			applicationId: 'a-1031',
		});
		await waitFor(() => expect(screen.getByLabelText('Тема')).toHaveValue(''));
	});

	it('при ошибке не теряет текст', async () => {
		server.use(http.post(apiUrl('/support/tickets'), () => HttpResponse.json(
			{ code: 'INTERNAL', message: 'Сервис недоступен' },
			{ status: 500 },
		)));
		const { user } = renderWithProviders(<CreateTicket />);

		await user.type(screen.getByLabelText('Тема'), 'Вопрос по ставке');
		await user.type(screen.getByLabelText('Сообщение'), 'Почему ставка выше ожидаемой?');
		await user.click(screen.getByRole('button', { name: 'Отправить обращение' }));

		expect(await screen.findByText('Сервис недоступен')).toBeInTheDocument();
		expect(screen.getByLabelText('Тема')).toHaveValue('Вопрос по ставке');
	});
});
