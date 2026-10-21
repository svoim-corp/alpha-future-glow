import { loanFixture } from '@test/fixtures/loans';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { CreditsPage } from './CreditsPage';

describe('CreditsPage', () => {
	it('показывает кредиты и ссылку на календарь', async () => {
		server.use(http.get(apiUrl('/loans'), () => HttpResponse.json([loanFixture])));
		renderWithProviders(<CreditsPage />);

		expect(screen.getByRole('heading', { level: 1, name: 'Мои кредиты' })).toBeInTheDocument();
		expect(await screen.findByRole('heading', { name: 'Оборотный кредит' })).toBeInTheDocument();
		expect(screen.getByRole('link', { name: 'Платёжный календарь' })).toHaveAttribute('href', '/credits/calendar');
	});
});
