import { counterOfferFixture, rejectedOfferFixture } from '@test/fixtures/tracking';
import { renderWithProviders } from '@test/render';
import { screen } from '@testing-library/react';

import { type OfferApproved, type OfferRejected } from '../model/tracking';
import { OfferView } from './OfferView';

describe('OfferView', () => {
	it('иные условия показывают запрошенное рядом с предложенным и причину', () => {
		renderWithProviders(<OfferView offer={counterOfferFixture} />);

		expect(screen.getByText('Сумма снижена по уровню долговой нагрузки')).toBeInTheDocument();
		expect(screen.getByRole('region', { name: 'Вы запрашивали' })).toHaveTextContent('1 500 000');
		expect(screen.getByRole('region', { name: 'Предложение банка' })).toHaveTextContent('1 200 000');
		expect(screen.getByText(/действует до 27\.10\.2026/)).toBeInTheDocument();
	});

	it('одобрение не показывает блок «запрашивали»', () => {
		const approved: OfferApproved = { ...(counterOfferFixture as OfferApproved), decision: 'APPROVED' };
		renderWithProviders(<OfferView offer={approved} />);

		expect(screen.queryByRole('region', { name: 'Вы запрашивали' })).not.toBeInTheDocument();
		expect(screen.getByText('Одобрено')).toBeInTheDocument();
	});

	it('отказ без альтернатив (бэкенд не присылает поле) не падает', () => {
		const { alternatives: _alternatives, ...withoutAlternatives } = rejectedOfferFixture as OfferRejected;
		renderWithProviders(<OfferView offer={withoutAlternatives} />);

		expect(screen.getByText('Банк не может одобрить заявку')).toBeInTheDocument();
		expect(screen.queryByText('Что можно сделать:')).not.toBeInTheDocument();
	});

	it('отказ называет причину и показывает альтернативы', () => {
		renderWithProviders(<OfferView offer={rejectedOfferFixture} />);

		expect(screen.getByText('Высокая долговая нагрузка по данным БКИ')).toBeInTheDocument();
		expect(screen.getByText(/Экспресс-кредит — до 500\s000/)).toBeInTheDocument();
	});
});
