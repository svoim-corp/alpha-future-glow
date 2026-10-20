import { fullApplication } from '@test/fixtures/applications';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { FillFullApplication } from './FillFullApplication';

const submitted = { id: 'a-1031', number: 'КЗ-2026-001031', status: 'SUBMITTED', redirect: '/applications/a-1031' };

const renderForm = (application = fullApplication, onSubmitted = vi.fn()) => renderWithProviders(
	<FillFullApplication application={application} onSubmitted={onSubmitted} />,
);

describe('FillFullApplication', () => {
	it('для ООО показывает блок поручителя, для ИП нет', () => {
		const { rerender } = renderForm();
		expect(screen.getByLabelText('ИНН поручителя')).toBeInTheDocument();

		const ie = { ...fullApplication, company: { ...fullApplication.company!, type: 'IE' as const } };
		rerender(<FillFullApplication application={ie} onSubmitted={vi.fn()} />);
		expect(screen.queryByLabelText('ИНН поручителя')).not.toBeInTheDocument();
	});

	it('отдельное поле адреса появляется, если адрес не совпадает с регистрацией', async () => {
		const { user } = renderForm();

		expect(screen.queryByLabelText('Фактический адрес')).not.toBeInTheDocument();
		await user.click(screen.getByRole('checkbox', { name: 'Совпадает с адресом регистрации' }));

		expect(screen.getByLabelText('Фактический адрес')).toBeInTheDocument();
	});

	it('без подтверждения достоверности не отправляет и подсвечивает пропуск', async () => {
		const submit = vi.fn();
		server.use(
			http.post(apiUrl('/applications/a-1031/validate'), () => HttpResponse.json({ valid: true, missing: [] })),
			http.post(apiUrl('/applications/a-1031/submit'), () => {
				submit();

				return HttpResponse.json(submitted);
			}),
		);
		const { user } = renderForm();

		await user.click(screen.getByRole('button', { name: 'Отправить заявку' }));

		expect(await screen.findByText('Подтвердите достоверность сведений')).toBeInTheDocument();
		expect(submit).not.toHaveBeenCalled();
	});

	it('показывает незаполненное из validate', async () => {
		server.use(http.post(apiUrl('/applications/a-1031/validate'), () => HttpResponse.json({
			valid: false,
			missing: [{ field: 'stage2.guarantor.inn', message: 'Укажите ИНН поручителя' }],
		})));
		const { user } = renderForm();

		await user.click(screen.getByRole('checkbox', { name: 'Подтверждаю, что все сведения достоверны' }));
		await user.click(screen.getByRole('button', { name: 'Отправить заявку' }));

		expect(await screen.findByText('Укажите ИНН поручителя')).toBeInTheDocument();
	});

	it('отправляет заявку с If-Match и Idempotency-Key и сообщает адрес перехода', async () => {
		const headers = vi.fn();
		const onSubmitted = vi.fn();
		server.use(
			http.post(apiUrl('/applications/a-1031/validate'), () => HttpResponse.json({ valid: true, missing: [] })),
			http.post(apiUrl('/applications/a-1031/submit'), ({ request }) => {
				headers(request.headers.get('If-Match'), request.headers.get('Idempotency-Key'));

				return HttpResponse.json(submitted);
			}),
		);
		const { user } = renderForm(fullApplication, onSubmitted);

		await user.click(screen.getByRole('checkbox', { name: 'Подтверждаю, что все сведения достоверны' }));
		await user.click(screen.getByRole('button', { name: 'Отправить заявку' }));

		await waitFor(() => expect(onSubmitted).toHaveBeenCalledWith('/applications/a-1031'));
		expect(headers).toHaveBeenCalledWith('7', expect.stringMatching(/^[0-9a-f-]{36}$/u));
	});

	it('согласие поручителя фиксируется отдельным запросом', async () => {
		const spy = vi.fn();
		server.use(http.post(apiUrl('/applications/a-1031/guarantor/consent-bki'), () => {
			spy();

			return HttpResponse.json({ given: true, at: '2026-10-20T10:20:00Z' }, { status: 201 });
		}));
		const pending = {
			...fullApplication,
			stage2: {
				...fullApplication.stage2,
				guarantor: { ...fullApplication.stage2.guarantor!, bkiConsent: false },
			},
		};
		const { user } = renderForm(pending);

		await user.click(screen.getByRole('checkbox', { name: 'Поручитель согласен на запрос его кредитной истории' }));

		await waitFor(() => expect(spy).toHaveBeenCalled());
	});

	it('ошибка сервера при отправке не теряет введённое', async () => {
		server.use(
			http.post(apiUrl('/applications/a-1031/validate'), () => HttpResponse.json({ valid: true, missing: [] })),
			http.post(apiUrl('/applications/a-1031/submit'), () => HttpResponse.json(
				{ code: 'INVALID_TRANSITION', message: 'Заявка уже отправлена' },
				{ status: 409 },
			)),
		);
		const { user } = renderForm();

		await user.click(screen.getByRole('checkbox', { name: 'Подтверждаю, что все сведения достоверны' }));
		await user.click(screen.getByRole('button', { name: 'Отправить заявку' }));

		expect(await screen.findByText('Заявка уже отправлена')).toBeInTheDocument();
		expect(screen.getByLabelText('ИНН поручителя')).toHaveValue('771234567890');
	});
});
