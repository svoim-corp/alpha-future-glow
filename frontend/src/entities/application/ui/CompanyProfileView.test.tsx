import { companyProfileFixture } from '@test/fixtures/companyProfile';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { CompanyProfileView } from './CompanyProfileView';

describe('CompanyProfileView', () => {
	it('показывает данные реестров', async () => {
		server.use(http.get(apiUrl('/applications/a-1/company-profile'), () => HttpResponse.json(companyProfileFixture)));
		renderWithProviders(<CompanyProfileView applicationId="a-1" />);

		expect(await screen.findByText('г. Москва, ул. Пекарская, 7')).toBeInTheDocument();
		expect(screen.getByText('10.71 Производство хлеба')).toBeInTheDocument();
		expect(screen.getByText('УСН «Доходы минус расходы»')).toBeInTheDocument();
		expect(screen.getByText('Учтена')).toBeInTheDocument();
	});

	it('при ошибке даёт повторить', async () => {
		server.use(
			http.get(apiUrl('/applications/a-1/company-profile'), () => HttpResponse.json(
				{ code: 'INTERNAL', message: 'Ошибка' },
				{ status: 500 },
			), { once: true }),
			http.get(apiUrl('/applications/a-1/company-profile'), () => HttpResponse.json(companyProfileFixture)),
		);
		const { user } = renderWithProviders(<CompanyProfileView applicationId="a-1" />);

		await user.click(await screen.findByRole('button', { name: 'Повторить' }));

		expect(await screen.findByText('Смирнов Олег Викторович', { selector: 'dd' })).toBeInTheDocument();
	});
});
