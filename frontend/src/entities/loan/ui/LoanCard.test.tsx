import { loanDetailsFixture, loanFixture } from '@test/fixtures/loans';
import { renderWithProviders } from '@test/render';
import { screen } from '@testing-library/react';

import { LoanCard } from './LoanCard';
import { PaymentsTable } from './PaymentsTable';

describe('LoanCard', () => {
	it('показывает остаток, условия и ближайший платёж', () => {
		renderWithProviders(<LoanCard loan={loanFixture} />);

		expect(screen.getByRole('heading', { name: 'Оборотный кредит' })).toBeInTheDocument();
		expect(screen.getByText('Действует')).toBeInTheDocument();
		expect(screen.getByText(/Ближайший платёж 20\.11\.2026/)).toBeInTheDocument();
	});

	it('закрытый кредит без платежа не показывает ближайший', () => {
		renderWithProviders(<LoanCard loan={{ ...loanFixture, status: 'CLOSED', nextPayment: null }} />);

		expect(screen.getByText('Закрыт')).toBeInTheDocument();
		expect(screen.queryByText(/Ближайший платёж/)).not.toBeInTheDocument();
	});
});

describe('PaymentsTable', () => {
	it('выводит строки графика со статусом текстом', () => {
		renderWithProviders(<PaymentsTable payments={loanDetailsFixture.payments} />);

		expect(screen.getAllByRole('row')).toHaveLength(3);
		expect(screen.getAllByText('Запланирован')).toHaveLength(2);
	});
});
