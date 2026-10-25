import { productsFixture } from '@test/fixtures/products';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { ProductCard } from './ProductCard';

const express = productsFixture[0]!;
const working = productsFixture[1]!;
const overdraft = productsFixture[2]!;

describe('ProductCard', () => {
	it('показывает диапазон суммы, срок и ставку', () => {
		render(<ProductCard product={working} onSelect={vi.fn()} />);

		expect(screen.getByRole('heading', { name: 'Оборотный кредит' })).toBeInTheDocument();
		expect(screen.getByText('300 тыс. — 10 млн ₽')).toBeInTheDocument();
		expect(screen.getByText('3–36 мес.')).toBeInTheDocument();
		expect(screen.getByText('20–28 % годовых')).toBeInTheDocument();
	});

	it('клик по «Посмотреть условия» отдаёт код', async () => {
		const onSelect = vi.fn();
		render(<ProductCard product={express} onSelect={onSelect} />);

		await userEvent.click(screen.getByRole('button', { name: 'Условия: Экспресс-кредит' }));

		expect(onSelect).toHaveBeenCalledWith('EXPRESS');
	});

	it('овердрафт: вместо суммы — лимит', () => {
		render(<ProductCard product={overdraft} onSelect={vi.fn()} />);

		expect(screen.getByText('Лимит до 5 млн ₽')).toBeInTheDocument();
		expect(screen.getByText('Лимит на 12 месяцев')).toBeInTheDocument();
	});
});

