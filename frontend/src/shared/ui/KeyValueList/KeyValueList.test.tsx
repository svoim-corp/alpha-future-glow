import { render, screen } from '@testing-library/react';

import { KeyValueList } from './KeyValueList';

describe('KeyValueList', () => {
	it('показывает пары «название — значение»', () => {
		render(<KeyValueList items={[{ label: 'Сумма', value: '1 500 000 ₽' }, { label: 'Срок', value: '24 месяца' }]} />);

		expect(screen.getByText('Сумма')).toBeInTheDocument();
		expect(screen.getByText('24 месяца')).toBeInTheDocument();
	});
});
