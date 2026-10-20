import { fullApplication } from '@test/fixtures/applications';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { EditDraftParams } from './EditDraftParams';

const patch = (spy: (body: unknown, ifMatch: string | null) => void, result = {}) => http.patch(
	apiUrl('/applications/a-1031'),
	async ({ request }) => {
		spy(await request.json(), request.headers.get('If-Match'));

		return HttpResponse.json({ version: 8, savedAt: '2026-10-20T10:13:00Z', prelimInvalidated: false, ...result });
	},
);

describe('EditDraftParams', () => {
	it('сохраняет изменённую сумму на уходе из поля, отправляя версию в If-Match', async () => {
		const spy = vi.fn();
		server.use(patch(spy));
		const { user } = renderWithProviders(<EditDraftParams application={fullApplication} />);

		await user.clear(screen.getByLabelText('Сумма, ₽'));
		await user.type(screen.getByLabelText('Сумма, ₽'), '2000000');
		await user.tab();

		await waitFor(() => expect(spy).toHaveBeenCalledTimes(1));
		expect(spy).toHaveBeenCalledWith({ params: { amount: 2000000 } }, '7');
	});

	it('не отправляет запрос, если значения не менялись', async () => {
		const spy = vi.fn();
		server.use(patch(spy));
		const { user } = renderWithProviders(<EditDraftParams application={fullApplication} />);

		await user.click(screen.getByLabelText('Сумма, ₽'));
		await user.tab();

		expect(spy).not.toHaveBeenCalled();
	});

	it('показывает ошибку валидации и ничего не сохраняет', async () => {
		const spy = vi.fn();
		server.use(patch(spy));
		const { user } = renderWithProviders(<EditDraftParams application={fullApplication} />);

		await user.clear(screen.getByLabelText('Срок, мес.'));
		await user.type(screen.getByLabelText('Срок, мес.'), 'abc');
		await user.tab();

		expect(await screen.findByText('Срок: введите число месяцев')).toBeInTheDocument();
		expect(spy).not.toHaveBeenCalled();
	});

	it('предупреждает, что предварительное решение нужно получить заново', async () => {
		server.use(patch(vi.fn(), { prelimInvalidated: true }));
		const { user } = renderWithProviders(<EditDraftParams application={fullApplication} />);

		await user.clear(screen.getByLabelText('Срок, мес.'));
		await user.type(screen.getByLabelText('Срок, мес.'), '36');
		await user.tab();

		expect(await screen.findByText('Предварительное решение нужно получить заново.')).toBeInTheDocument();
	});

	it('при конфликте версий просит обновить страницу', async () => {
		server.use(http.patch(apiUrl('/applications/a-1031'), () => HttpResponse.json(
			{ code: 'VERSION_CONFLICT', message: 'Версия устарела' },
			{ status: 412 },
		)));
		const { user } = renderWithProviders(<EditDraftParams application={fullApplication} />);

		await user.clear(screen.getByLabelText('Сумма, ₽'));
		await user.type(screen.getByLabelText('Сумма, ₽'), '900000');
		await user.tab();

		expect(await screen.findByText('Заявка изменена в другом окне')).toBeInTheDocument();
	});

	it('блокирует поля, когда параметры зафиксированы', () => {
		const locked = { ...fullApplication, locks: { paramsLocked: true, stage2Available: true } };
		renderWithProviders(<EditDraftParams application={locked} />);

		expect(screen.getByLabelText('Сумма, ₽')).toBeDisabled();
	});

	it('показывает поле отсрочки только для инвестиционного кредита', () => {
		const invest = { ...fullApplication, product: { code: 'INVEST', name: 'Инвестиционный кредит' } };
		renderWithProviders(<EditDraftParams application={invest} />);

		expect(screen.getByLabelText('Отсрочка основного долга, мес.')).toBeInTheDocument();
	});
});
