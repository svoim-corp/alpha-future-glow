import { queueItemFixture } from '@test/fixtures/bank';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { AssignApplication } from './AssignApplication';

describe('AssignApplication', () => {
	it('берёт заявку в работу', async () => {
		const spy = vi.fn();
		server.use(http.post(apiUrl('/bank/applications/a-1024/assign'), () => {
			spy();

			return HttpResponse.json({ assignee: { id: 'u-9', fullName: 'Иванова Е. П.' } });
		}));
		const { user } = renderWithProviders(<AssignApplication item={queueItemFixture} />);

		await user.click(screen.getByRole('button', { name: 'Взять в работу КЗ-2026-001024' }));

		await waitFor(() => expect(spy).toHaveBeenCalled());
	});

	it('если заявку уже взяли, сообщает об этом', async () => {
		server.use(http.post(apiUrl('/bank/applications/a-1024/assign'), () => HttpResponse.json(
			{ code: 'CONFLICT', message: 'Заявка уже взята' },
			{ status: 409 },
		)));
		const { user } = renderWithProviders(<AssignApplication item={queueItemFixture} />);

		await user.click(screen.getByRole('button', { name: 'Взять в работу КЗ-2026-001024' }));

		expect(await screen.findByText('Заявку уже взял другой сотрудник')).toBeInTheDocument();
	});

	it('для занятой заявки показывает исполнителя без кнопки', () => {
		const taken = { ...queueItemFixture, assignee: { id: 'u-9', fullName: 'Иванова Е. П.' } };
		renderWithProviders(<AssignApplication item={taken} />);

		expect(screen.getByText('Иванова Е. П.')).toBeInTheDocument();
		expect(screen.queryByRole('button')).not.toBeInTheDocument();
	});
});
