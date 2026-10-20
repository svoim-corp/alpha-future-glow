import { scheduleFixture } from '@test/fixtures/loans';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { PaymentCalendar } from './PaymentCalendar';

describe('PaymentCalendar', () => {
	beforeEach(() => {
		vi.useFakeTimers({ toFake: ['Date'] });
		vi.setSystemTime(new Date('2026-11-05T09:00:00Z'));
	});

	afterEach(() => {
		vi.useRealTimers();
	});

	it('показывает платёж в дне месяца и список ближайших', async () => {
		server.use(
			http.get(apiUrl('/loans/schedule'), () => HttpResponse.json(scheduleFixture.slice(0, 1))),
			http.get(apiUrl('/loans/schedule/upcoming'), () => HttpResponse.json(scheduleFixture)),
		);
		renderWithProviders(<PaymentCalendar />);

		expect(await screen.findByText(/Платёж 62\s200/)).toBeInTheDocument();
		expect(screen.getByRole('heading', { name: 'ноябрь 2026' })).toBeInTheDocument();
		expect(await screen.findByRole('list', { name: 'Ближайшие платежи' })).toBeInTheDocument();
	});

	it('переключение месяца запрашивает расписание нового периода', async () => {
		const from = vi.fn();
		server.use(
			http.get(apiUrl('/loans/schedule'), ({ request }) => {
				from(new URL(request.url).searchParams.get('from'));

				return HttpResponse.json([]);
			}),
			http.get(apiUrl('/loans/schedule/upcoming'), () => HttpResponse.json([])),
		);
		const { user } = renderWithProviders(<PaymentCalendar />);

		await user.click(await screen.findByRole('button', { name: 'Следующий месяц' }));

		await screen.findByRole('heading', { name: 'декабрь 2026' });
		expect(from).toHaveBeenCalledWith('2026-11-01');
		expect(from).toHaveBeenCalledWith('2026-12-01');
	});

	it('если ближайших платежей нет, говорит об этом', async () => {
		server.use(
			http.get(apiUrl('/loans/schedule'), () => HttpResponse.json([])),
			http.get(apiUrl('/loans/schedule/upcoming'), () => HttpResponse.json([])),
		);
		renderWithProviders(<PaymentCalendar />);

		expect(await screen.findByText('Ближайших платежей нет')).toBeInTheDocument();
	});
});
