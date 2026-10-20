import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { Button } from './Button';

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

	it('в состоянии loading заблокирована и не вызывает onClick', async () => {
		const onClick = vi.fn();
		render(<Button loading onClick={onClick}>Отправить</Button>);

		const button = screen.getByRole('button');
		await userEvent.click(button);

		expect(button).toBeDisabled();
		expect(button).toHaveAttribute('aria-busy', 'true');
		expect(onClick).not.toHaveBeenCalled();
	});
});
