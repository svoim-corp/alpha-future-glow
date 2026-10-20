import { fullApplication } from '@test/fixtures/applications';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { VerifyCompany } from './VerifyCompany';

const empty = { ...fullApplication, company: null, authority: null };
const lookupBody = {
	company: {
		inn: '7712345678',
		type: 'LLC',
		name: 'ООО «Северная пекарня»',
		ogrn: '1157746000000',
		regDate: '2015-03-12',
		state: 'ACTIVE',
		head: 'Смирнов Олег Викторович',
	},
	applicantIsHead: false,
	authorityRequired: 'MCHD',
};

describe('VerifyCompany', () => {
	it('показывает карточку компании, уже подтянутую в заявку', () => {
		renderWithProviders(<VerifyCompany application={fullApplication} />);

		expect(screen.getByLabelText('ИНН компании или ИП')).toHaveValue('7712345678');
		expect(screen.getByText('ООО «Северная пекарня»')).toBeInTheDocument();
		expect(screen.queryByText('Вы подаёте заявку как представитель')).not.toBeInTheDocument();
	});

	it('проверяет ИНН на месте и не отправляет запрос с ошибкой', async () => {
		const { user } = renderWithProviders(<VerifyCompany application={empty} />);

		await user.type(screen.getByLabelText('ИНН компании или ИП'), '123');
		await user.click(screen.getByRole('button', { name: 'Найти компанию' }));

		expect(await screen.findByText(/10 цифр для организации или 12 для ИП/)).toBeInTheDocument();
	});

	it('для представителя просит МЧД и принимает действующую', async () => {
		server.use(
			http.post(apiUrl('/applications/a-1031/company'), () => HttpResponse.json(lookupBody)),
			http.post(apiUrl('/applications/a-1031/authority'), () => HttpResponse.json({
				verified: true,
				validUntil: '2027-03-01',
			})),
		);
		const { user } = renderWithProviders(<VerifyCompany application={empty} />);

		await user.type(screen.getByLabelText('ИНН компании или ИП'), '7712345678');
		await user.click(screen.getByRole('button', { name: 'Найти компанию' }));
		expect(await screen.findByText('Вы подаёте заявку как представитель')).toBeInTheDocument();
		expect(screen.getByText(/Смирнов Олег Викторович/)).toBeInTheDocument();

		await user.type(screen.getByLabelText('Номер машиночитаемой доверенности'), 'a1b2c3');
		await user.click(screen.getByRole('button', { name: 'Проверить доверенность' }));

		expect(await screen.findByText('Полномочия подтверждены')).toBeInTheDocument();
		expect(screen.getByText(/до 01\.03\.2027/)).toBeInTheDocument();
	});

	it('недействительная МЧД объясняет причину и напоминает, что подать может руководитель', async () => {
		server.use(
			http.post(apiUrl('/applications/a-1031/company'), () => HttpResponse.json(lookupBody)),
			http.post(apiUrl('/applications/a-1031/authority'), () => HttpResponse.json({
				verified: false,
				reason: 'Срок действия истёк',
			})),
		);
		const { user } = renderWithProviders(<VerifyCompany application={empty} />);

		await user.type(screen.getByLabelText('ИНН компании или ИП'), '7712345678');
		await user.click(screen.getByRole('button', { name: 'Найти компанию' }));
		await user.type(await screen.findByLabelText('Номер машиночитаемой доверенности'), 'a1b2c3');
		await user.click(screen.getByRole('button', { name: 'Проверить доверенность' }));

		expect(await screen.findByText('Доверенность не подошла')).toBeInTheDocument();
		expect(screen.getByText(/Срок действия истёк\. Подать заявку может только руководитель\./)).toBeInTheDocument();
	});

	it('если компания не найдена в реестре, показывает сообщение сервера', async () => {
		server.use(http.post(apiUrl('/applications/a-1031/company'), () => HttpResponse.json(
			{ code: 'NOT_FOUND', message: 'Компания с таким ИНН не найдена' },
			{ status: 404 },
		)));
		const { user } = renderWithProviders(<VerifyCompany application={empty} />);

		await user.type(screen.getByLabelText('ИНН компании или ИП'), '7700000000');
		await user.click(screen.getByRole('button', { name: 'Найти компанию' }));

		expect(await screen.findByText('Компания с таким ИНН не найдена')).toBeInTheDocument();
	});

	it('при сбое проверки МЧД показывает ошибку', async () => {
		server.use(http.post(apiUrl('/applications/a-1031/authority'), () => HttpResponse.json(
			{ code: 'UPSTREAM_UNAVAILABLE', message: 'Реестр МЧД недоступен' },
			{ status: 502 },
		)));
		const app = { ...fullApplication, authority: { type: 'MCHD' as const, verified: false } };
		const { user } = renderWithProviders(<VerifyCompany application={app} />);

		await user.type(screen.getByLabelText('Номер машиночитаемой доверенности'), 'a1b2c3');
		await user.click(screen.getByRole('button', { name: 'Проверить доверенность' }));

		expect(await screen.findByText('Реестр МЧД недоступен')).toBeInTheDocument();
	});
});
