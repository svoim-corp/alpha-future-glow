import { makePlan } from '@test/fixtures/plans';
import { investDetails } from '@test/fixtures/products';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { PlanDetails } from './PlanDetails';
import { PlanFunding } from './PlanFunding';

const plan = makePlan({ recommendation: {
	productCode: 'INVEST', productName: 'Инвестиционный кредит', reason: 'Подходит по цели и сумме', availableFrom: '2026-10-21',
} });

beforeEach(() => {
	server.use(http.get(apiUrl('/products/INVEST'), () => HttpResponse.json(investDetails)));
});

describe('краткое описание финансирования плана', () => {
	it('оставляет описание и неактивный переход, без срока и требований', async () => {
		const create = vi.fn();
		server.use(http.post(apiUrl('/applications'), () => { create(); return HttpResponse.json({}); }));
		const { user } = renderWithProviders(<PlanFunding productCode="INVEST" productName="Инвестиционный кредит" />);
		expect(await screen.findByText(investDetails.description)).toBeInTheDocument();
		expect(screen.queryByLabelText('Срок кредита, месяцев')).not.toBeInTheDocument();
		expect(screen.queryByText('Что понадобится')).not.toBeInTheDocument();
		const button = screen.getByRole('button', { name: 'Перейти к оформлению' });
		expect(button).toBeDisabled();
		expect(button).toHaveAccessibleDescription('Страница оформления пока недоступна.');
		await user.click(button);
		expect(create).not.toHaveBeenCalled();
	});

	it('показывает недоступность описания и умеет повторить загрузку', async () => {
		server.use(http.get(apiUrl('/products/INVEST'), () => new HttpResponse(null, { status: 501 }), { once: true }));
		const { user } = renderWithProviders(<PlanFunding productCode="INVEST" productName="Инвестиционный кредит" />);
		expect(await screen.findByRole('alert')).toHaveTextContent('Описание кредита пока недоступно');
		await user.click(screen.getByRole('button', { name: 'Повторить загрузку' }));
		expect(await screen.findByText(investDetails.description)).toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'Перейти к оформлению' })).toBeDisabled();
	});

	it('для будущей рекомендации тоже оставляет только описание и отключённый переход', async () => {
		renderWithProviders(<PlanDetails plan={{ ...plan, recommendation: {
			productCode: 'INVEST', productName: 'Инвестиционный кредит', reason: 'Подходит', availableFrom: '2027-01-01',
		} }} today="2026-10-21" onClose={vi.fn()} />);
		expect(await screen.findByText(investDetails.description)).toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'Перейти к оформлению' })).toBeDisabled();
	});

	it.each(['PAUSED', 'DONE', 'ARCHIVED'] as const)('не подменяет состояние %s рекомендацией', (state) => {
		renderWithProviders(<PlanDetails plan={{ ...plan, state }} today="2026-10-21" onClose={vi.fn()} />);
		expect(screen.getByText('В этом состоянии план не участвует в подборе.')).toBeInTheDocument();
	});
});
