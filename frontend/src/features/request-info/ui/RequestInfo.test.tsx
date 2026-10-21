import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen, waitFor, within } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { RequestInfo } from './RequestInfo';

const open = async (user: ReturnType<typeof renderWithProviders>['user']) => {
	await user.click(screen.getByRole('button', { name: 'Запросить информацию' }));

	return screen.getByRole('dialog');
};

describe('RequestInfo', () => {
	it('требует комментарий и хотя бы один вопрос', async () => {
		const { user } = renderWithProviders(<RequestInfo applicationId="a-1" />);
		const dialog = await open(user);

		await user.click(within(dialog).getByRole('button', { name: 'Отправить запрос' }));

		expect(await within(dialog).findByText('Напишите комментарий и хотя бы один вопрос клиенту')).toBeInTheDocument();
	});

	it('каждая непустая строка становится пунктом запроса', async () => {
		const body = vi.fn();
		server.use(http.post(apiUrl('/bank/applications/a-1/info-requests'), async ({ request }) => {
			body(await request.json());

			return HttpResponse.json({ id: 'r-7', status: 'OPEN', applicationStatus: 'INFO_REQUESTED' }, { status: 201 });
		}));
		const { user } = renderWithProviders(<RequestInfo applicationId="a-1" />);
		const dialog = await open(user);

		await user.type(within(dialog).getByLabelText('Комментарий клиенту'), 'Расхождение оборотов');
		await user.type(within(dialog).getByLabelText('Вопросы клиенту'), 'Адрес склада{Enter}{Enter}Причина расхождения');
		await user.click(within(dialog).getByRole('button', { name: 'Отправить запрос' }));

		await waitFor(() => expect(body).toHaveBeenCalledWith({
			comment: 'Расхождение оборотов',
			items: [
				{ key: 'item_1', type: 'TEXT', label: 'Адрес склада' },
				{ key: 'item_2', type: 'TEXT', label: 'Причина расхождения' },
			],
		}));
		await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
	});
});
