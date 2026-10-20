import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen, waitFor, within } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { DecideOffer } from './DecideOffer';

describe('DecideOffer', () => {
	it('принятие условий передаёт адрес подписания', async () => {
		const onAccepted = vi.fn();
		server.use(http.post(apiUrl('/applications/a-1/offer/accept'), () => HttpResponse.json({
			status: 'SIGNING',
			redirect: '/applications/a-1/sign',
		})));
		const { user } = renderWithProviders(<DecideOffer applicationId="a-1" onAccepted={onAccepted} />);

		await user.click(screen.getByRole('button', { name: 'Принять условия' }));

		await waitFor(() => expect(onAccepted).toHaveBeenCalledWith('/applications/a-1/sign'));
	});

	it('отказ требует подтверждения и отправляет причину', async () => {
		const body = vi.fn();
		server.use(http.post(apiUrl('/applications/a-1/offer/decline'), async ({ request }) => {
			body(await request.json());

			return HttpResponse.json({ status: 'CLOSED' });
		}));
		const { user } = renderWithProviders(<DecideOffer applicationId="a-1" onAccepted={vi.fn()} />);

		await user.click(screen.getByRole('button', { name: 'Отказаться' }));
		expect(body).not.toHaveBeenCalled();
		await user.type(screen.getByLabelText('Причина (необязательно)'), 'Нашли другое финансирование');
		await user.click(within(screen.getByRole('dialog')).getByRole('button', { name: 'Отказаться' }));

		await waitFor(() => expect(body).toHaveBeenCalledWith({ reason: 'Нашли другое финансирование' }));
		await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
	});

	it('показывает ошибку, если принять не удалось', async () => {
		server.use(http.post(apiUrl('/applications/a-1/offer/accept'), () => HttpResponse.json(
			{ code: 'INVALID_TRANSITION', message: 'Срок предложения истёк' },
			{ status: 409 },
		)));
		const { user } = renderWithProviders(<DecideOffer applicationId="a-1" onAccepted={vi.fn()} />);

		await user.click(screen.getByRole('button', { name: 'Принять условия' }));

		expect(await screen.findByText('Срок предложения истёк')).toBeInTheDocument();
	});
});
