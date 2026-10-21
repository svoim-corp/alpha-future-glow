import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { AsyncBoundary } from './AsyncBoundary';

const base = { isLoading: false, isError: false, onRetry: vi.fn() };

describe('AsyncBoundary', () => {
	it('при загрузке показывает status', () => {
		render(<AsyncBoundary {...base} isLoading>Данные</AsyncBoundary>);

		expect(screen.getByRole('status', { name: 'Загрузка' })).toBeInTheDocument();
		expect(screen.queryByText('Данные')).not.toBeInTheDocument();
	});

	it('при ошибке показывает alert и умеет повторить', async () => {
		const onRetry = vi.fn();
		render(<AsyncBoundary {...base} isError onRetry={onRetry}>Данные</AsyncBoundary>);

		expect(screen.getByRole('alert')).toHaveTextContent('Не удалось загрузить данные.');
		await userEvent.click(screen.getByRole('button', { name: 'Повторить' }));

		expect(onRetry).toHaveBeenCalledOnce();
	});

	it('при пустом результате показывает пояснение и действие', () => {
		render(
			<AsyncBoundary {...base} isEmpty emptyTitle="Заявок нет" emptyText="Начните с каталога" emptyAction={<a href="/products">Каталог</a>}>
				Данные
			</AsyncBoundary>,
		);

		expect(screen.getByText('Заявок нет')).toBeInTheDocument();
		expect(screen.getByText('Начните с каталога')).toBeInTheDocument();
		expect(screen.getByRole('link', { name: 'Каталог' })).toBeInTheDocument();
	});

	it('при наличии данных показывает содержимое', () => {
		render(<AsyncBoundary {...base}>Данные</AsyncBoundary>);

		expect(screen.getByText('Данные')).toBeInTheDocument();
	});
});
