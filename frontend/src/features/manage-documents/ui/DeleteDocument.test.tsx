import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen, waitFor, within } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { DeleteDocument } from './DeleteDocument';

describe('DeleteDocument', () => {
	it('удаляет только после подтверждения', async () => {
		const spy = vi.fn();
		server.use(http.delete(apiUrl('/documents/doc-1'), () => {
			spy();

			return new HttpResponse(null, { status: 204 });
		}));
		const { user } = renderWithProviders(<DeleteDocument id="doc-1" name="Реквизиты компании" />);

		await user.click(screen.getByRole('button', { name: 'Удалить: Реквизиты компании' }));
		expect(spy).not.toHaveBeenCalled();
		await user.click(within(screen.getByRole('dialog')).getByRole('button', { name: 'Удалить' }));

		await waitFor(() => expect(spy).toHaveBeenCalledTimes(1));
		await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
	});

	it('отмена ничего не удаляет', async () => {
		const { user } = renderWithProviders(<DeleteDocument id="doc-1" name="Реквизиты компании" />);

		await user.click(screen.getByRole('button', { name: 'Удалить: Реквизиты компании' }));
		await user.click(screen.getByRole('button', { name: 'Отмена' }));

		expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
	});
});
