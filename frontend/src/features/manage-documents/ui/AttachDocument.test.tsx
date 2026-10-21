import { draftApplication, infoRequestedApplication } from '@test/fixtures/applications';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen, waitFor, within } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { AttachDocument } from './AttachDocument';

const list = (items: unknown[]) => http.get(apiUrl('/applications'), () => HttpResponse.json({
	items,
	page: 1,
	size: 20,
	total: items.length,
}));

describe('AttachDocument', () => {
	it('предлагает только черновики и прикрепляет файл к выбранному', async () => {
		const body = vi.fn();
		server.use(
			list([draftApplication, infoRequestedApplication]),
			http.post(apiUrl('/applications/a-1031/documents'), async ({ request }) => {
				body(await request.json());

				return new HttpResponse(null, { status: 204 });
			}),
		);
		const { user } = renderWithProviders(<AttachDocument id="doc-1" name="Реквизиты" />);

		await user.click(screen.getByRole('button', { name: 'В заявку: Реквизиты' }));
		const dialog = await screen.findByRole('dialog');
		expect(await within(dialog).findAllByRole('radio')).toHaveLength(1);
		await user.click(within(dialog).getByRole('radio'));
		await user.click(within(dialog).getByRole('button', { name: 'Прикрепить' }));

		await waitFor(() => expect(body).toHaveBeenCalledWith({ documentId: 'doc-1' }));
		expect(await screen.findByText('Файл прикреплён к черновику')).toBeInTheDocument();
	});

	it('без черновиков объясняет, почему прикрепить некуда', async () => {
		server.use(list([infoRequestedApplication]));
		const { user } = renderWithProviders(<AttachDocument id="doc-1" name="Реквизиты" />);

		await user.click(screen.getByRole('button', { name: 'В заявку: Реквизиты' }));

		expect(await screen.findByText('Нет черновиков')).toBeInTheDocument();
	});
});
