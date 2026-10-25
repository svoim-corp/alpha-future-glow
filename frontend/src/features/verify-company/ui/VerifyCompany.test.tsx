import { fullApplication } from '@test/fixtures/applications';
import { renderWithProviders } from '@test/render';
import { screen } from '@testing-library/react';

import { VerifyCompany } from './VerifyCompany';

describe('VerifyCompany', () => {
	it('показывает ИНН из аккаунта в заблокированном поле и карточку компании', () => {
		renderWithProviders(<VerifyCompany application={fullApplication} />);

		const inn = screen.getByLabelText('ИНН компании или ИП');
		expect(inn).toHaveValue('7712345678');
		expect(inn).toBeDisabled();
		expect(screen.getByText('ООО «Северная пекарня»')).toBeInTheDocument();
		expect(screen.queryByRole('button')).not.toBeInTheDocument();
		expect(screen.queryByText('Вы подаёте заявку как представитель')).not.toBeInTheDocument();
	});

	it('для представителя по МЧД показывает срок действия доверенности', () => {
		const application = {
			...fullApplication,
			authority: { type: 'MCHD' as const, verified: true, validUntil: '2027-03-01' },
		};
		renderWithProviders(<VerifyCompany application={application} />);

		expect(screen.getByText('Полномочия подтверждены')).toBeInTheDocument();
		expect(screen.getByText(/Доверенность действует до/)).toBeInTheDocument();
	});

	it('без данных компании показывает пустой ИНН и не падает', () => {
		renderWithProviders(<VerifyCompany application={{ ...fullApplication, company: null, authority: null }} />);

		expect(screen.getByLabelText('ИНН компании или ИП')).toHaveValue('');
	});
});
