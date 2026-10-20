import { productsFixture } from '@test/fixtures/products';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { ProductCard } from './ProductCard';

const express = productsFixture[0]!;
const working = productsFixture[1]!;
const overdraft = productsFixture[2]!;

describe('ProductCard', () => {
	it('показывает ключевые условия и порог без выписки', () => {
		render(<ProductCard product={working} selected={false} fits={false} onSelect={vi.fn()} />);

		expect(screen.getByRole('heading', { name: 'Оборотный кредит' })).toBeInTheDocument();
		expect(screen.getByText(/до 10\s000\s000/)).toBeInTheDocument();
		expect(screen.getByText('3–36 мес.')).toBeInTheDocument();
		expect(screen.getByText(/Без выписки до/)).toBeInTheDocument();
	});

	it('подходящий продукт помечен, выбранный — aria-pressed, клик отдаёт код', async () => {
		const onSelect = vi.fn();
		render(<ProductCard product={express} selected fits onSelect={onSelect} estimate={<p>Платёж ≈ 1 ₽</p>} />);

		const button = screen.getByRole('button', { name: 'Условия: Экспресс-кредит', pressed: true });
		await userEvent.click(button);

		expect(screen.getByText('Подходит под задачу')).toBeInTheDocument();
		expect(screen.getByText('Платёж ≈ 1 ₽')).toBeInTheDocument();
		expect(onSelect).toHaveBeenCalledWith('EXPRESS');
	});

	it('овердрафт: вместо суммы — лимит, нет срока, помечен для клиентов банка', () => {
		render(<ProductCard product={overdraft} selected={false} fits={false} onSelect={vi.fn()} />);

		expect(screen.getByText('Лимит')).toBeInTheDocument();
		expect(screen.queryByText('Срок')).not.toBeInTheDocument();
		expect(screen.getByText('Только клиентам банка')).toBeInTheDocument();
	});
});
