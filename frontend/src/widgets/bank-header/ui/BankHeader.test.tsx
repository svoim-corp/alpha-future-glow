import { officerUser } from '@test/fixtures/session';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen } from '@testing-library/react';
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
});
