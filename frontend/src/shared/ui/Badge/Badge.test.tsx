import { render, screen } from '@testing-library/react';

import { Card } from '../Card';
import { Badge } from './Badge';

describe('Badge', () => {
	it('показывает подпись статуса', () => {
		render(<Badge tone="success">Кредит выдан</Badge>);

		expect(screen.getByText('Кредит выдан')).toBeInTheDocument();
	});

	it('нейтральный вариант без тона', () => {
		render(<Badge>Черновик</Badge>);

		expect(screen.getByText('Черновик')).toBeInTheDocument();
	});
});

describe('Card', () => {
	it('по умолчанию — секция, можно сменить тег и сделать акцентной', () => {
		render(<><Card aria-label="Условия">a</Card><Card as="article" accent>b</Card></>);

		expect(screen.getByRole('region', { name: 'Условия' })).toBeInTheDocument();
		expect(screen.getByRole('article')).toBeInTheDocument();
	});
});
