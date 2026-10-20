import { render, screen } from '@testing-library/react';

import { Page } from './Page';

describe('Page', () => {
	it('рендерит заголовок первого уровня, пояснение, действия и содержимое', () => {
		render(<Page title="Мои заявки" subtitle="Статусы и черновики" actions={<button type="button">Создать</button>}>Тело</Page>);

		expect(screen.getByRole('heading', { level: 1, name: 'Мои заявки' })).toBeInTheDocument();
		expect(screen.getByText('Статусы и черновики')).toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'Создать' })).toBeInTheDocument();
		expect(screen.getByText('Тело')).toBeInTheDocument();
	});

	it('без пояснения и действий отображает только заголовок', () => {
		render(<Page title="Поддержка">Тело</Page>);

		expect(screen.getByRole('main')).toBeInTheDocument();
	});
});
