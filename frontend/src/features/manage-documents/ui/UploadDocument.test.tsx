import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen, waitFor, within } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { UploadDocument } from './UploadDocument';

const pdf = () => new File(['%PDF'], 'rekvizity.pdf', { type: 'application/pdf' });

describe('UploadDocument', () => {
	it('загружает файл с выбранной категорией и закрывает окно', async () => {
		const spy = vi.fn();
		server.use(http.post(apiUrl('/documents'), async ({ request }) => {
			const form = await request.formData();
			spy(typeof form.get('file') !== 'string', form.get('category'));

			return HttpResponse.json({ id: 'doc-9' }, { status: 201 });
		}));
		const { user } = renderWithProviders(<UploadDocument />);

		await user.click(screen.getByRole('button', { name: 'Добавить документ' }));
		const dialog = screen.getByRole('dialog');
		await user.upload(within(dialog).getByLabelText(/Файл/), pdf());
		await user.selectOptions(within(dialog).getByLabelText('Категория'), 'CONTRACTS');
		await user.click(within(dialog).getByRole('button', { name: 'Загрузить' }));

		await waitFor(() => expect(spy).toHaveBeenCalledWith(true, 'CONTRACTS'));
		await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
	});

	it('без файла загрузка недоступна', async () => {
		const { user } = renderWithProviders(<UploadDocument />);

		await user.click(screen.getByRole('button', { name: 'Добавить документ' }));

		expect(within(screen.getByRole('dialog')).getByRole('button', { name: 'Загрузить' })).toBeDisabled();
	});

	it('ошибка сервера остаётся в окне', async () => {
		server.use(http.post(apiUrl('/documents'), () => HttpResponse.json(
			{ code: 'PAYLOAD_TOO_LARGE', message: 'Файл слишком большой' },
			{ status: 413 },
		)));
		const { user } = renderWithProviders(<UploadDocument />);

		await user.click(screen.getByRole('button', { name: 'Добавить документ' }));
		const dialog = screen.getByRole('dialog');
		await user.upload(within(dialog).getByLabelText(/Файл/), pdf());
		await user.click(within(dialog).getByRole('button', { name: 'Загрузить' }));

		expect(await within(dialog).findByText('Файл слишком большой')).toBeInTheDocument();
	});
});
