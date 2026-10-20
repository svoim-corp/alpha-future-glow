import { render, screen } from '@testing-library/react';

import { Steps } from './Steps';

describe('Steps', () => {
	it('отмечает текущий шаг через aria-current', () => {
		render(<Steps label="Этапы заявки" labels={['Предварительное решение', 'Полная заявка']} current={1} />);

		const items = screen.getAllByRole('listitem');

		expect(screen.getByRole('list', { name: 'Этапы заявки' })).toBeInTheDocument();
		expect(items[1]).toHaveAttribute('aria-current', 'step');
		expect(items[0]).not.toHaveAttribute('aria-current');
	});
});
