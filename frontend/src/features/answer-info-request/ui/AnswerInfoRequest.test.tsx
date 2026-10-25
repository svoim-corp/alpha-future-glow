import { openInfoRequest } from '@test/fixtures/tracking';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { AnswerInfoRequest } from './AnswerInfoRequest';

const request = openInfoRequest;
const render = () => renderWithProviders(<AnswerInfoRequest applicationId="a-1024" request={request} />);

describe('AnswerInfoRequest', () => {
	it('не отправляет, пока не заполнены все пункты', async () => {
		const spy = vi.fn();
		server.use(http.post(apiUrl('/applications/a-1024/info-requests/r-7/answer'), () => {
			spy();

			return HttpResponse.json({ requestStatus: 'ANSWERED', applicationStatus: 'DATA_CHECK' });
		}));
		const { user } = render();

		await user.type(screen.getByLabelText('Адрес склада, где хранится товар'), 'г. Москва');
		await user.click(screen.getByRole('button', { name: 'Отправить ответ' }));

		expect(await screen.findByText('Заполните ответ')).toBeInTheDocument();
		expect(spy).not.toHaveBeenCalled();
	});

	it('отправляет ответы по ключам пунктов', async () => {
		const body = vi.fn();
		server.use(http.post(apiUrl('/applications/a-1024/info-requests/r-7/answer'), async ({ request: req }) => {
			body(await req.json());

			return HttpResponse.json({ requestStatus: 'ANSWERED', applicationStatus: 'DATA_CHECK' });
		}));
		const { user } = render();

		await user.type(screen.getByLabelText('Адрес склада, где хранится товар'), 'г. Москва, ул. Складская, 3');
		await user.type(screen.getByLabelText(/Почему выручка/), 'Возврат займа');
		await user.click(screen.getByRole('button', { name: 'Отправить ответ' }));

		await waitFor(() => expect(body).toHaveBeenCalledWith({
			answers: [
				{ key: 'actual_address', value: 'г. Москва, ул. Складская, 3' },
				{ key: 'turnover_note', value: 'Возврат займа' },
			],
		}));
	});

	it('после отправки поля становятся только для чтения', async () => {
		server.use(http.post(apiUrl('/applications/a-1024/info-requests/r-7/answer'), () => HttpResponse.json(
			{ requestStatus: 'ANSWERED', applicationStatus: 'DATA_CHECK' },
		)));
		const { user } = render();

		await user.type(screen.getByLabelText('Адрес склада, где хранится товар'), 'Адрес');
		await user.type(screen.getByLabelText(/Почему выручка/), 'Причина');
		await user.click(screen.getByRole('button', { name: 'Отправить ответ' }));

		expect(await screen.findByText('Ответ отправлен банку')).toBeInTheDocument();
		expect(screen.getByLabelText('Адрес склада, где хранится товар')).toHaveAttribute('readonly');
		expect(screen.queryByRole('button', { name: 'Отправить ответ' })).not.toBeInTheDocument();
	});

	it('если на запрос уже ответили, показывает ответы только для чтения', () => {
		renderWithProviders(
			<AnswerInfoRequest
				applicationId="a-1024"
				request={{
					...request,
					status: 'ANSWERED',
					answers: [
						{ key: 'actual_address', value: 'г. Москва' },
						{ key: 'turnover_note', value: 'Возврат займа' },
					],
				}}
			/>,
		);

		expect(screen.getByLabelText('Адрес склада, где хранится товар')).toHaveValue('г. Москва');
		expect(screen.getByLabelText('Адрес склада, где хранится товар')).toHaveAttribute('readonly');
		expect(screen.queryByRole('button', { name: 'Отправить ответ' })).not.toBeInTheDocument();
	});

	it('на конфликт «уже ответили» блокирует поля, а не показывает ошибку', async () => {
		server.use(http.post(apiUrl('/applications/a-1024/info-requests/r-7/answer'), () => HttpResponse.json(
			{ code: 'INVALID_TRANSITION', message: 'На этот запрос уже ответили' },
			{ status: 409 },
		)));
		const { user } = render();

		await user.type(screen.getByLabelText('Адрес склада, где хранится товар'), 'Адрес');
		await user.type(screen.getByLabelText(/Почему выручка/), 'Причина');
		await user.click(screen.getByRole('button', { name: 'Отправить ответ' }));

		expect(await screen.findByText('Ответ отправлен банку')).toBeInTheDocument();
		expect(screen.getByLabelText('Адрес склада, где хранится товар')).toHaveValue('Адрес');
		expect(screen.queryByText('На этот запрос уже ответили')).not.toBeInTheDocument();
	});

	it('при ошибке сохраняет введённый текст', async () => {
		server.use(http.post(apiUrl('/applications/a-1024/info-requests/r-7/answer'), () => HttpResponse.json(
			{ code: 'INTERNAL', message: 'Сервис недоступен' },
			{ status: 500 },
		)));
		const { user } = render();

		await user.type(screen.getByLabelText('Адрес склада, где хранится товар'), 'Адрес');
		await user.type(screen.getByLabelText(/Почему выручка/), 'Причина');
		await user.click(screen.getByRole('button', { name: 'Отправить ответ' }));

		expect(await screen.findByText('Сервис недоступен')).toBeInTheDocument();
		expect(screen.getByLabelText('Адрес склада, где хранится товар')).toHaveValue('Адрес');
	});
});
