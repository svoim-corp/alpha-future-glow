import { render, screen } from '@testing-library/react';

import { Alert } from './Alert';

describe('Alert', () => {
	it('ошибка объявляется как alert', () => {
		render(<Alert tone="danger" title="Не удалось сохранить">Повторите попытку</Alert>);

		expect(screen.getByRole('alert')).toHaveTextContent('Не удалось сохранить');
	});

	it('информационное сообщение — status, действия отображаются', () => {
		render(<Alert title="Черновик сохранён" actions={<button type="button">Ок</button>} />);

		expect(screen.getByRole('status')).toHaveTextContent('Черновик сохранён');
		expect(screen.getByRole('button', { name: 'Ок' })).toBeInTheDocument();
	});
});
