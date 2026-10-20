import { contractFixture } from '@test/fixtures/contract';
import { renderWithProviders } from '@test/render';
import { screen } from '@testing-library/react';

import { ContractTermsView } from './ContractTerms';

describe('ContractTermsView', () => {
	it('показывает условия и ссылки на документы', () => {
		renderWithProviders(<ContractTermsView contract={contractFixture} />);

		expect(screen.getByText('Ежемесячный платёж')).toBeInTheDocument();
		expect(screen.getByText('20.11.2026')).toBeInTheDocument();
		expect(screen.getByRole('link', { name: 'Кредитный договор' })).toHaveAttribute('href', '/api/v1/documents/d-1');
		expect(screen.getAllByRole('listitem')).toHaveLength(2);
	});
});
