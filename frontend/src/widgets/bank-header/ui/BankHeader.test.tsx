import { officerUser } from '@test/fixtures/session';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';

import { BankHeader } from './BankHeader';

describe('BankHeader', () => {
	it('ведёт в очередь заявок и показывает сотрудника с должностью', async () => {
		server.use(http.get(apiUrl('/auth/me'), () => HttpResponse.json(officerUser)));
		renderWithProviders(<BankHeader />);

		expect(screen.getByRole('link', { name: /Кредитный кабинет/ })).toHaveAttribute('href', '/bank/queue');
		expect(await screen.findByText(/Иванова Елена Петровна, Кредитный инспектор/)).toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'Выйти' })).toBeInTheDocument();
	});

	it('переключает тёмную тему и применяет её к странице', async () => {
		server.use(http.get(apiUrl('/auth/me'), () => HttpResponse.json(officerUser)));
		localStorage.removeItem('credit-ui-theme');
		renderWithProviders(<BankHeader />);
		const toggle = screen.getByRole('button', { name: 'Тёмная тема' });
		const before = toggle.getAttribute('aria-pressed');

		await userEvent.click(toggle);

		expect(toggle).toHaveAttribute('aria-pressed', before === 'true' ? 'false' : 'true');
		expect(document.documentElement.dataset.theme).toBe(before === 'true' ? 'light' : 'dark');
	});
});
