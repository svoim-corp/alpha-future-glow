import { fullApplication } from '@test/fixtures/applications';
import { render, screen } from '@testing-library/react';

import { CompanyCard } from './CompanyCard';

const company = fullApplication.company!;

describe('CompanyCard', () => {
	it('показывает название, ОГРН, дату регистрации и состояние', () => {
		render(<CompanyCard company={company} />);

		expect(screen.getByText('ООО «Северная пекарня»')).toBeInTheDocument();
		expect(screen.getByText(/ОГРН 1157746000000 · с 12\.03\.2015/)).toBeInTheDocument();
		expect(screen.getByText('действует')).toBeInTheDocument();
		expect(screen.getByText('ЕГРЮЛ')).toBeInTheDocument();
	});

	it('для ИП указывает ЕГРИП и ОГРНИП', () => {
		render(<CompanyCard company={{ ...company, type: 'IE', state: 'LIQUIDATING' }} />);

		expect(screen.getByText('ЕГРИП')).toBeInTheDocument();
		expect(screen.getByText(/ОГРНИП/)).toBeInTheDocument();
		expect(screen.getByText('в процессе ликвидации')).toBeInTheDocument();
	});
});
