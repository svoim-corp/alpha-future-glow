import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { isVersionConflict, useDraftPatch } from './useDraftPatch';

const Saver = () => {
	const { save, isSaving } = useDraftPatch('a-1', 7);

	return (
		<>
			<button
				type="button"
				onClick={() => {
					void save({ params: { amount: 1 } }).catch(() => undefined);
					void save({ params: { amount: 2 } }).catch(() => undefined);
				}}
			>
				Сохранить дважды
			</button>
			<p>{isSaving ? 'сохраняем' : 'простой'}</p>
		</>
	);
};

describe('useDraftPatch', () => {
	it('сохраняет по очереди, подставляя новую версию из предыдущего ответа', async () => {
		const versions: string[] = [];
		server.use(http.patch(apiUrl('/applications/a-1'), ({ request }) => {
			const current = Number(request.headers.get('If-Match'));
			versions.push(String(current));

			return HttpResponse.json({ version: current + 1, savedAt: '2026-10-20T10:13:00Z', prelimInvalidated: false });
		}));
		const { user } = renderWithProviders(<Saver />);

		await user.click(screen.getByRole('button', { name: 'Сохранить дважды' }));

		await waitFor(() => expect(versions).toEqual(['7', '8']));
	});

	it('после ошибки очередь продолжает работать', async () => {
		let calls = 0;
		server.use(http.patch(apiUrl('/applications/a-1'), () => {
			calls += 1;

			return calls === 1
				? HttpResponse.json({ code: 'VERSION_CONFLICT', message: 'Изменено' }, { status: 412 })
				: HttpResponse.json({ version: 9, savedAt: '2026-10-20T10:13:00Z', prelimInvalidated: false });
		}));
		const { user } = renderWithProviders(<Saver />);

		await user.click(screen.getByRole('button', { name: 'Сохранить дважды' }));

		await waitFor(() => expect(calls).toBe(2));
	});
});

describe('isVersionConflict', () => {
	it('узнаёт VERSION_CONFLICT по коду ошибки', () => {
		expect(isVersionConflict({ status: 412, data: { code: 'VERSION_CONFLICT', message: 'x' } })).toBe(true);
		expect(isVersionConflict({ status: 400, data: { code: 'VALIDATION_ERROR', message: 'x' } })).toBe(false);
		expect(isVersionConflict(undefined)).toBe(false);
	});
});
