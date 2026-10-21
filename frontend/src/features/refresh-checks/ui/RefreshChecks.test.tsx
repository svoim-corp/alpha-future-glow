import { checksFixture } from '@test/fixtures/bank';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { RefreshChecks } from './RefreshChecks';

describe('RefreshChecks', () => {
	it('перезапрашивает проверки', async () => {
		const spy = vi.fn();
		server.use(http.post(apiUrl('/bank/applications/a-1/checks/refresh'), () => {
			spy();

			return HttpResponse.json(checksFixture);
		}));
		const { user } = renderWithProviders(<RefreshChecks applicationId="a-1" />);

		await user.click(screen.getByRole('button', { name: 'Обновить проверки' }));

		await waitFor(() => expect(spy).toHaveBeenCalled());
	});

	it('при сбое сообщает об ошибке', async () => {
		server.use(http.post(apiUrl('/bank/applications/a-1/checks/refresh'), () => HttpResponse.json(
			{ code: 'UPSTREAM_UNAVAILABLE', message: 'Источники недоступны' },
			{ status: 502 },
		)));
		const { user } = renderWithProviders(<RefreshChecks applicationId="a-1" />);

		await user.click(screen.getByRole('button', { name: 'Обновить проверки' }));

		expect(await screen.findByText('Источники недоступны')).toBeInTheDocument();
	});
});
