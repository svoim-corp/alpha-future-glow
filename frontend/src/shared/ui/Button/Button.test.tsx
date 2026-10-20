import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router';

import { Button, ButtonLink } from './index';

describe('Button', () => {
	it('вызывает onClick по клику', async () => {
		const onClick = vi.fn();
		render(<Button onClick={onClick}>Отправить</Button>);

		await userEvent.click(screen.getByRole('button', { name: 'Отправить' }));

		expect(onClick).toHaveBeenCalledOnce();
	});

	it('по умолчанию имеет type="button", чтобы не отправлять форму случайно', () => {
		render(<Button>Ок</Button>);

		expect(screen.getByRole('button')).toHaveAttribute('type', 'button');
	});

	it('в состоянии loading заблокирована, сохраняет подпись и не вызывает onClick', async () => {
		const onClick = vi.fn();
		render(<Button loading onClick={onClick}>Отправить</Button>);

		const button = screen.getByRole('button', { name: 'Отправить' });
		await userEvent.click(button);

		expect(button).toBeDisabled();
		expect(button).toHaveAttribute('aria-busy', 'true');
		expect(onClick).not.toHaveBeenCalled();
	});

	it('поддерживает все варианты и компактный размер', () => {
		render(
			<>
				<Button variant="secondary">Вторичная</Button>
				<Button variant="ghost">Прозрачная</Button>
				<Button variant="danger" compact>Опасная</Button>
			</>,
		);

		expect(screen.getAllByRole('button')).toHaveLength(3);
	});
});

describe('ButtonLink', () => {
	it('рендерит ссылку на маршрут', () => {
		render(<MemoryRouter><ButtonLink to="/products" variant="secondary" compact>Каталог</ButtonLink></MemoryRouter>);

		expect(screen.getByRole('link', { name: 'Каталог' })).toHaveAttribute('href', '/products');
	});
});
