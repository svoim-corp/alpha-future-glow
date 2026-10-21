import { homeRouteFor } from './homeRoute';

describe('homeRouteFor', () => {
	it('клиент попадает на главную, сотрудник банка — в очередь заявок', () => {
		expect(homeRouteFor('CLIENT')).toBe('/');
		expect(homeRouteFor('BANK_OFFICER')).toBe('/bank/queue');
	});
});
