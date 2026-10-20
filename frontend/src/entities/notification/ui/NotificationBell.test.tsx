import { notificationsFixture } from '@test/fixtures/notifications';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { Route, Routes } from 'react-router';

import { NotificationBell } from './NotificationBell';

const renderBell = () => renderWithProviders(
	<Routes>
		<Route path="/" element={<NotificationBell />} />
		<Route path="/applications/:id" element={<p>Карточка заявки</p>} />
	</Routes>,
);

describe('NotificationBell', () => {
	it('показывает число непрочитанных в имени кнопки', async () => {
		server.use(http.get(apiUrl('/notifications'), () => HttpResponse.json(notificationsFixture)));
		renderBell();

		expect(await screen.findByRole('button', { name: 'Уведомления, непрочитанных: 1' })).toBeInTheDocument();
	});

	it('клик по уведомлению отмечает его прочитанным и открывает ссылку', async () => {
		const read = vi.fn();
		server.use(
			http.get(apiUrl('/notifications'), () => HttpResponse.json(notificationsFixture)),
			http.post(apiUrl('/notifications/n-15/read'), () => {
				read();

				return new HttpResponse(null, { status: 204 });
			}),
		);
		const { user } = renderBell();

		await user.click(await screen.findByRole('button', { name: /Уведомления/ }));
		await user.click(screen.getByRole('button', { name: /Банк просит дополнить заявку/ }));

		expect(await screen.findByText('Карточка заявки')).toBeInTheDocument();
		await waitFor(() => expect(read).toHaveBeenCalledTimes(1));
	});

	it('прочитанное уведомление не отправляет повторный запрос', async () => {
		const read = vi.fn();
		server.use(
			http.get(apiUrl('/notifications'), () => HttpResponse.json({
				unreadCount: 0,
				items: [{ ...notificationsFixture.items[0]!, read: true }],
			})),
			http.post(apiUrl('/notifications/n-15/read'), () => {
				read();

				return new HttpResponse(null, { status: 204 });
			}),
		);
		const { user } = renderBell();

		await user.click(await screen.findByRole('button', { name: 'Уведомления' }));
		await user.click(screen.getByRole('button', { name: /Банк просит дополнить заявку/ }));

		expect(await screen.findByText('Карточка заявки')).toBeInTheDocument();
		expect(read).not.toHaveBeenCalled();
	});

	it('Escape закрывает список', async () => {
		server.use(http.get(apiUrl('/notifications'), () => HttpResponse.json(notificationsFixture)));
		const { user } = renderBell();

		await user.click(await screen.findByRole('button', { name: /Уведомления/ }));
		expect(screen.getByRole('region', { name: 'Список уведомлений' })).toBeInTheDocument();
		await user.keyboard('{Escape}');

		expect(screen.queryByRole('region', { name: 'Список уведомлений' })).not.toBeInTheDocument();
	});

	it('пустой список объясняет, что новых нет', async () => {
		server.use(http.get(apiUrl('/notifications'), () => HttpResponse.json({ unreadCount: 0, items: [] })));
		const { user } = renderBell();

		await user.click(await screen.findByRole('button', { name: 'Уведомления' }));

		expect(screen.getByText('Новых уведомлений нет')).toBeInTheDocument();
	});
});
