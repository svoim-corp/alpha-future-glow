import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { ConfirmSupportAction } from './ConfirmSupportAction';

describe('Подтверждение действия в чате', () => {
	it('выполняет команду только по кнопке и показывает результат', async () => {
		const confirm = vi.fn();
		server.use(http.post(apiUrl('/support/conversations/chat-1/actions/action-1/confirm'), ({ request }) => {
			confirm(request.headers.get('Idempotency-Key'));
			return HttpResponse.json({ kind: 'result', title: 'Обращение отправлено', resultId: 't-1' });
		}));
		const { user } = renderWithProviders(<ConfirmSupportAction conversationId="chat-1" actionId="action-1"
			action={{ kind: 'ticket', subject: 'Не открывается заявка', text: 'Ошибка при переходе в заявку' }} />);
		expect(confirm).not.toHaveBeenCalled();
		await user.click(screen.getByRole('button', { name: 'Подтвердить' }));
		expect(await screen.findByText('Обращение отправлено')).toBeInTheDocument();
		expect(screen.getByText('Номер обращения: t-1')).toBeInTheDocument();
		expect(confirm).toHaveBeenCalledOnce();
		expect(confirm).toHaveBeenCalledWith('action-1');
		expect(screen.queryByRole('button', { name: 'Подтвердить' })).not.toBeInTheDocument();
	});

	it('при отказе сервера позволяет повторить подтверждение', async () => {
		server.use(http.post(apiUrl('/support/conversations/chat-1/actions/action-1/confirm'), () =>
			HttpResponse.json({ code: 'BUSINESS_RULE', message: 'Продукт недоступен' }, { status: 422 })));
		const { user } = renderWithProviders(<ConfirmSupportAction conversationId="chat-1" actionId="action-1"
			action={{ kind: 'draft', draft: { productCode: 'INVEST', amount: 2000000, term: 24, goal: 'EQUIPMENT' } }} />);
		await user.click(screen.getByRole('button', { name: 'Подтвердить' }));
		expect(await screen.findByText('Продукт недоступен')).toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'Подтвердить' })).toBeEnabled();
	});
});
