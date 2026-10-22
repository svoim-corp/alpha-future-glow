import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { useState } from 'react';

import { Select } from '../Select';
import { Modal } from './Modal';

const Host = () => {
	const [open, setOpen] = useState(false);

	return (
		<>
			<button type="button" onClick={() => setOpen(true)}>Открыть</button>
			{open && (
				<Modal title="Подтверждение подписи" onClose={() => setOpen(false)} footer={<button type="button">Подписать</button>}>
					<input aria-label="Поле" />
				</Modal>
			)}
		</>
	);
};

describe('Modal', () => {
	it('открывается как диалог с заголовком и фокусом внутри', async () => {
		render(<Host />);
		await userEvent.click(screen.getByRole('button', { name: 'Открыть' }));

		const dialog = screen.getByRole('dialog', { name: 'Подтверждение подписи' });

		expect(dialog).toHaveAttribute('aria-modal', 'true');
		expect(dialog).toHaveFocus();
	});

	it('Escape закрывает диалог и возвращает фокус на кнопку открытия', async () => {
		render(<Host />);
		const opener = screen.getByRole('button', { name: 'Открыть' });
		await userEvent.click(opener);

		await userEvent.keyboard('{Escape}');

		expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
		expect(opener).toHaveFocus();
	});

	it('кнопка «Закрыть» закрывает диалог', async () => {
		render(<Host />);
		await userEvent.click(screen.getByRole('button', { name: 'Открыть' }));

		await userEvent.click(screen.getByRole('button', { name: 'Закрыть' }));

		expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
	});

	it('Tab не выводит фокус за пределы диалога', async () => {
		render(<Host />);
		await userEvent.click(screen.getByRole('button', { name: 'Открыть' }));

		await userEvent.tab();
		await userEvent.tab();
		await userEvent.tab();
		await userEvent.tab();

		expect(screen.getByRole('dialog')).toContainElement(document.activeElement as HTMLElement);
	});

	it('Escape сначала закрывает меню выбора, затем само окно', async () => {
		const close = vi.fn();
		const user = userEvent.setup();
		render(<Modal title="План" onClose={close}>
			<Select aria-label="Состояние" value="ACTIVE" onChange={vi.fn()}
				options={[{ value: 'ACTIVE', label: 'Активен' }]} />
		</Modal>);
		await user.click(screen.getByRole('combobox'));
		expect(screen.getByRole('listbox')).toBeInTheDocument();
		await user.keyboard('{Escape}');
		expect(screen.queryByRole('listbox')).not.toBeInTheDocument();
		expect(close).not.toHaveBeenCalled();
		await user.keyboard('{Escape}');
		expect(close).toHaveBeenCalledOnce();
	});
});
