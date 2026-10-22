import { type ChatBlock } from '@entities/support';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { ChatBlocks } from './ChatBlocks';

describe('Карточки возможностей ассистента', () => {
	it('предлагает удалить только черновик и выполняет команду после подтверждения', async () => {
		const deleted = vi.fn();
		server.use(http.delete(apiUrl('/applications/draft-1'), () => {
			deleted();
			return new HttpResponse(null, { status: 204 });
		}));
		const application = {
			id: 'draft-1', number: null, productCode: 'EXPRESS', amount: 150000,
			term: 12, status: 'Черновик', nextStep: 'Заполнить', canDelete: true,
		};
		const { user } = renderWithProviders(<ChatBlocks conversationId="chat-1" blocks={[
			{ kind: 'application', title: 'Черновик', application },
			{ kind: 'application', title: 'Заявка', application: {
				...application, id: 'sent-1', number: '1', status: 'Отправлена', nextStep: 'Ожидайте', canDelete: false,
			} },
		]} />);
		expect(screen.getAllByRole('button', { name: 'Удалить черновик' })).toHaveLength(1);
		await user.click(screen.getByRole('button', { name: 'Удалить черновик' }));
		expect(deleted).not.toHaveBeenCalled();
		await user.click(screen.getByRole('button', { name: /^Удалить$/ }));
		await waitFor(() => expect(deleted).toHaveBeenCalledOnce());
	});
	it('даёт настоящую ссылку на планер и показывает проверенные данные', () => {
		const blocks: ChatBlock[] = [
			{ kind: 'navigation', title: 'Планер', navigation: { section: 'planner', label: 'Перейти в планер', href: '/planner' } },
			{ kind: 'plan', title: 'Купить печь', plan: {
				id: 'p-1', title: 'Купить печь', goal: 'EQUIPMENT', budgetRub: 500000, dateMode: 'DAY',
				startDate: '2026-11-10', state: 'ACTIVE', updatedAt: '2026-10-22T12:00:00Z',
			} },
			{ kind: 'company', title: 'Пекарня', company: {
				name: 'Северная пекарня', inn: '7712345678', regAddress: 'Адрес регистрации',
				actualAddress: 'Адрес склада', monthlyIncome: 1000000,
			} },
			{ kind: 'loan', title: 'Кредит', loan: {
				id: 'l-1', productName: 'Оборотный кредит', amount: 1000000, balance: 500000, rate: 20, term: 24, status: 'ACTIVE',
			} },
			{ kind: 'ticket', title: 'Не открывается заявка', ticket: {
				id: 't-1', subject: 'Не открывается заявка', text: 'Ошибка перехода', status: 'OPEN',
			} },
		];
		renderWithProviders(<ChatBlocks blocks={blocks} conversationId="chat-1" />);
		expect(screen.getByRole('link', { name: 'Перейти в планер' })).toHaveAttribute('href', '/planner');
		expect(screen.getByRole('heading', { name: 'Купить печь' })).toBeInTheDocument();
		expect(screen.getByText('Купить оборудование')).toBeInTheDocument();
		expect(screen.getByText('7712345678')).toBeInTheDocument();
		expect(screen.getByText('Остаток основного долга')).toBeInTheDocument();
		expect(screen.getByText('Рассматривается')).toBeInTheDocument();
		expect(screen.queryByRole('button', { name: 'Подтвердить' })).not.toBeInTheDocument();
	});
});
