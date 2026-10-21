import { chooseOption } from '@test/select';
import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { useState } from 'react';

import { Modal } from '../Modal';
import { DateField } from './DateField';

const DateInput = () => {
	const [value, setValue] = useState('2026-10-21');

	return <><DateField label="Дата" value={value} onValueChange={setValue} /><output aria-label="ISO">{value}</output></>;
};

describe('DateField на DayPicker', () => {
	it('показывает русскую дату и сохраняет ручной ввод в ISO', async () => {
		const user = userEvent.setup();
		render(<DateInput />);
		expect(screen.getByLabelText('Дата')).toHaveValue('21.10.2026');
		await user.clear(screen.getByLabelText('Дата'));
		await user.type(screen.getByLabelText('Дата'), '29.02.2028');
		expect(screen.getByLabelText('ISO')).toHaveTextContent('2028-02-29');
		await user.clear(screen.getByLabelText('Дата'));
		await user.type(screen.getByLabelText('Дата'), '31.02.2028');
		expect(screen.getByLabelText('ISO')).toHaveTextContent('31.02.2028');
	});

	it('выбирает день в библиотечном календаре с выбором месяца и года', async () => {
		const user = userEvent.setup();
		render(<DateInput />);
		await user.click(screen.getByRole('button', { name: 'Выбрать дату: Дата' }));
		expect(screen.getByRole('group', { name: 'Выбор даты' })).toBeInTheDocument();
		await chooseOption(user, 'Месяц', 'февраль');
		await chooseOption(user, 'Год', '2028');
		const calendar = within(screen.getByRole('group', { name: 'Выбор даты' }));
		await user.click(calendar.getByRole('button', { name: /29.*феврал.*2028/i }));
		expect(screen.getByLabelText('Дата')).toHaveValue('29.02.2028');
		expect(screen.getByLabelText('ISO')).toHaveTextContent('2028-02-29');
		expect(screen.queryByRole('group', { name: 'Выбор даты' })).not.toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'Выбрать дату: Дата' })).toHaveFocus();
	});

	it('Escape закрывает календарь внутри формы, сохраняя модальное окно', async () => {
		const close = vi.fn();
		const user = userEvent.setup();
		render(<Modal title="План" onClose={close}><DateInput /></Modal>);
		await user.click(screen.getByRole('button', { name: 'Выбрать дату: Дата' }));
		await user.click(screen.getByRole('combobox', { name: 'Месяц' }));
		await user.keyboard('{Escape}');
		expect(screen.queryByRole('listbox')).not.toBeInTheDocument();
		expect(screen.getByRole('group', { name: 'Выбор даты' })).toBeInTheDocument();
		await user.keyboard('{Escape}');
		expect(screen.queryByRole('group', { name: 'Выбор даты' })).not.toBeInTheDocument();
		expect(close).not.toHaveBeenCalled();
		await user.keyboard('{Escape}');
		expect(close).toHaveBeenCalledOnce();
	});

	it('не открывает календарь у заблокированного поля и показывает ошибку', async () => {
		const user = userEvent.setup();
		render(<DateField label="Дата" value="" onValueChange={vi.fn()} disabled error="Укажите дату" />);
		await user.click(screen.getByRole('button', { name: 'Выбрать дату: Дата' }));
		expect(screen.queryByRole('group', { name: 'Выбор даты' })).not.toBeInTheDocument();
		expect(screen.getByLabelText('Дата')).toHaveAccessibleDescription('Укажите дату');
	});
});
