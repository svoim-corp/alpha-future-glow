import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { DeleteDraftButton } from './DeleteDraftButton';

const setup = () => renderWithProviders(<DeleteDraftButton applicationId="a-1031" productName="Экспресс-кредит" />);

describe('DeleteDraftButton', () => {
	it('сначала просит подтверждение и не удаляет без него', async () => {
		let called = false;
		server.use(http.delete(apiUrl('/applications/a-1031'), () => {
			called = true;

			return new HttpResponse(null, { status: 204 });
		}));
		const { user } = setup();

		await user.click(screen.getByRole('button', { name: 'Удалить черновик' }));
		expect(screen.getByRole('dialog', { name: 'Удалить черновик?' })).toBeInTheDocument();

		await user.click(screen.getByRole('button', { name: 'Оставить' }));

		expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
		expect(called).toBe(false);
	});

	it('после подтверждения удаляет черновик и закрывает диалог', async () => {
		let called = false;
		server.use(http.delete(apiUrl('/applications/a-1031'), () => {
			called = true;

			return new HttpResponse(null, { status: 204 });
		}));
		const { user } = setup();

		await user.click(screen.getByRole('button', { name: 'Удалить черновик' }));
		await user.click(screen.getByRole('button', { name: 'Удалить' }));

		await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
		expect(called).toBe(true);
	});

	it('при ошибке сервера объясняет причину и оставляет диалог открытым', async () => {
		server.use(http.delete(apiUrl('/applications/a-1031'), () => HttpResponse.json(
			{ code: 'INVALID_TRANSITION', message: 'Удалить можно только черновик' },
			{ status: 409 },
		)));
		const { user } = setup();

		await user.click(screen.getByRole('button', { name: 'Удалить черновик' }));
		await user.click(screen.getByRole('button', { name: 'Удалить' }));

		expect(await screen.findByRole('alert')).toHaveTextContent('Удалить можно только черновик');
		expect(screen.getByRole('dialog')).toBeInTheDocument();
	});
});
