import { officerUser } from '@test/fixtures/session';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';

import { BankShell } from './BankShell';

describe('BankShell', () => {
	it('ведёт в очередь заявок и показывает сотрудника с должностью', async () => {
		server.use(http.get(apiUrl('/auth/me'), () => HttpResponse.json(officerUser)));
		renderWithProviders(<BankShell><p>Содержимое</p></BankShell>);

		expect(screen.getByRole('link', { name: /Кредитный\s*кабинет/ })).toHaveAttribute('href', '/bank/queue');
		expect(await screen.findByText('Елена Иванова')).toBeInTheDocument();
		expect(screen.getByText('Кредитный инспектор')).toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'Выйти' })).toBeInTheDocument();
	});

	it('переключает тёмную тему и применяет её к странице', async () => {
		server.use(http.get(apiUrl('/auth/me'), () => HttpResponse.json(officerUser)));
		localStorage.removeItem('credit-ui-theme');
		renderWithProviders(<BankShell><p>Содержимое</p></BankShell>);
		const toggle = screen.getByRole('button', { name: 'Тёмная тема' });
		const before = toggle.getAttribute('aria-pressed');

		await userEvent.click(toggle);

		expect(toggle).toHaveAttribute('aria-pressed', before === 'true' ? 'false' : 'true');
		expect(document.documentElement.dataset.theme).toBe(before === 'true' ? 'light' : 'dark');
	});
});
