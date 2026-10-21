import { assistantMessageFixture, conversationFixture } from '@test/fixtures/chat';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { SupportChat } from './SupportChat';

describe('Чат поддержки', () => {
	it('показывает пустой диалог, создаёт его и отправляет вопрос с ключом', async () => {
		const sent = vi.fn();
		server.use(
			http.post(apiUrl('/support/conversations'), () => HttpResponse.json(conversationFixture, { status: 201 })),
			http.get(apiUrl('/support/conversations/chat-1/messages'), () => HttpResponse.json([])),
			http.post(apiUrl('/support/conversations/chat-1/messages'), async ({ request }) => {
				sent(await request.json(), request.headers.get('Idempotency-Key'));
				return HttpResponse.json({ ...assistantMessageFixture, role: 'user' }, { status: 202 });
			}),
		);
		const { user } = renderWithProviders(<SupportChat />);
		expect(await screen.findByText('С чего начнём?')).toBeInTheDocument();
		await user.type(screen.getByRole('textbox', { name: 'Ваш вопрос' }), 'Покажи мои заявки');
		await user.click(screen.getByRole('button', { name: 'Отправить вопрос' }));
		await waitFor(() => expect(sent).toHaveBeenCalledWith({ text: 'Покажи мои заявки' }, expect.any(String)));
		await waitFor(() => expect(screen.getByRole('textbox', { name: 'Ваш вопрос' })).toHaveValue(''));
	});

	it('показывает реальные карточки и источники из истории', async () => {
		server.use(
			http.get(apiUrl('/support/conversations'), () => HttpResponse.json([conversationFixture])),
			http.get(apiUrl('/support/conversations/chat-1/messages'), () => HttpResponse.json([assistantMessageFixture])),
		);
		renderWithProviders(<SupportChat />);
		expect(await screen.findByText('Вот ближайший платёж.')).toBeInTheDocument();
		expect(screen.getByRole('heading', { name: 'Оборотный кредит' })).toBeInTheDocument();
		expect(screen.getByRole('link', { name: 'Открыть календарь платежей' })).toHaveAttribute('href', '/credits/calendar');
		expect(screen.getByText('Материалы по вопросу')).toBeInTheDocument();
	});

	it('при недоступности модели сохраняет введённый вопрос и ключ для повтора', async () => {
		const keys: (string | null)[] = [];
		server.use(
			http.get(apiUrl('/support/conversations'), () => HttpResponse.json([conversationFixture])),
			http.get(apiUrl('/support/conversations/chat-1/messages'), () => HttpResponse.json([])),
			http.post(apiUrl('/support/conversations/chat-1/messages'), ({ request }) => {
				keys.push(request.headers.get('Idempotency-Key'));
				return HttpResponse.json({ code: 'UPSTREAM_UNAVAILABLE', message: 'Модель недоступна' }, { status: 502 });
			}),
		);
		const { user } = renderWithProviders(<SupportChat />);
		await screen.findByText('С чего начнём?');
		await user.type(screen.getByRole('textbox', { name: 'Ваш вопрос' }), 'Как оформить кредит?');
		await user.click(screen.getByRole('button', { name: 'Отправить вопрос' }));
		expect(await screen.findByText('Модель недоступна')).toBeInTheDocument();
		expect(screen.getByRole('textbox', { name: 'Ваш вопрос' })).toHaveValue('Как оформить кредит?');
		await user.click(screen.getByRole('button', { name: 'Отправить вопрос' }));
		await waitFor(() => expect(keys).toHaveLength(2));
		expect(keys[0]).toBe(keys[1]);
	});

	it('после ошибки загрузки умеет повторить запрос', async () => {
		server.use(http.get(apiUrl('/support/conversations'), () => HttpResponse.json({}, { status: 503 }), { once: true }));
		const { user } = renderWithProviders(<SupportChat />);
		expect(await screen.findByText('Не удалось загрузить переписку')).toBeInTheDocument();
		await user.click(screen.getByRole('button', { name: 'Повторить' }));
		expect(await screen.findByText('С чего начнём?')).toBeInTheDocument();
	});

	it('ожидает ответ и блокирует отправку второго вопроса', async () => {
		server.use(
			http.get(apiUrl('/support/conversations'), () => HttpResponse.json([conversationFixture])),
			http.get(apiUrl('/support/conversations/chat-1/messages'), () => HttpResponse.json([
				{ ...assistantMessageFixture, status: 'PENDING', text: '', blocks: [], sources: [] },
			])),
		);
		renderWithProviders(<SupportChat />);
		expect(await screen.findByText('Ассистент готовит ответ…')).toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'Отправить вопрос' })).toBeDisabled();
	});
});
