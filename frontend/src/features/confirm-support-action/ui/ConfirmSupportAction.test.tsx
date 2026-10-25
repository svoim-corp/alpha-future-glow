import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { ConfirmSupportAction } from './ConfirmSupportAction';

describe('Подтверждение действия в чате', () => {
	it('сохраняет план только после подтверждения и показывает переход в планер', async () => {
		const called = vi.fn();
		const plan = {
			id: 'p-5', title: 'Новая печь', goal: 'EQUIPMENT', budgetRub: 500000, dateMode: 'DAY' as const,
			startDate: '2026-11-10', state: 'ACTIVE' as const, updatedAt: '2026-10-22T12:00:00Z',
		};
		server.use(http.post(apiUrl('/support/conversations/chat-1/actions/plan-action/confirm'), () => {
			called();
			return HttpResponse.json({ kind: 'result', title: 'План сохранён', resultId: 'p-5', plan });
		}));
		const { user } = renderWithProviders(<ConfirmSupportAction conversationId="chat-1" actionId="plan-action"
			action={{ kind: 'plan', planPreview: plan, planChange: { budgetRub: 500000 } }} />);
		expect(screen.getByRole('heading', { name: 'Новая печь' })).toBeInTheDocument();
		expect(called).not.toHaveBeenCalled();
		await user.click(screen.getByRole('button', { name: 'Подтвердить' }));
		expect(await screen.findByText('План сохранён')).toBeInTheDocument();
		expect(screen.getByRole('link', { name: 'Открыть сохранённый план в планере' })).toHaveAttribute('href', '/planner');
		expect(called).toHaveBeenCalledOnce();
	});
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
