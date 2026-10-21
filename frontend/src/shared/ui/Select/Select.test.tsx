import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { useState } from 'react';
import { describe, expect, it, vi } from 'vitest';

import { SelectField } from '../Field';
import { Select } from './index';

const OPTIONS = [
	{ value: 'usn', label: 'УСН' },
	{ value: 'patent', label: 'Патент' },
	{ value: 'osno', label: 'ОСНО' },
];

const Controlled = ({ onChange }: { onChange?: (value: string) => void }) => {
	const [value, setValue] = useState('usn');

	return (
		<Select
			aria-label="Режим"
			options={OPTIONS}
			value={value}
			onChange={(next) => {
				setValue(next);
				onChange?.(next);
			}}
		/>
	);
};

describe('Select', () => {
	it('показывает выбранное значение и открывает список по клику', async () => {
		const user = userEvent.setup();
		render(<Controlled />);

		const trigger = screen.getByRole('combobox', { name: 'Режим' });
		expect(trigger).toHaveTextContent('УСН');
		expect(trigger).toHaveAttribute('aria-expanded', 'false');

		await user.click(trigger);

		expect(trigger).toHaveAttribute('aria-expanded', 'true');
		expect(screen.getByRole('option', { name: 'УСН' })).toHaveAttribute('aria-selected', 'true');
		expect(screen.getAllByRole('option')).toHaveLength(3);
	});

	it('выбор кликом меняет значение и закрывает список', async () => {
		const user = userEvent.setup();
		const onChange = vi.fn();
		render(<Controlled onChange={onChange} />);

		await user.click(screen.getByRole('combobox'));
		await user.click(screen.getByRole('option', { name: 'Патент' }));

		expect(onChange).toHaveBeenCalledWith('patent');
		expect(screen.getByRole('combobox')).toHaveTextContent('Патент');
		expect(screen.queryByRole('listbox')).not.toBeInTheDocument();
	});

	it('работает с клавиатуры: стрелки, Enter, Escape', async () => {
		const user = userEvent.setup();
		const onChange = vi.fn();
		render(<Controlled onChange={onChange} />);

		screen.getByRole('combobox').focus();
		await user.keyboard('{ArrowDown}');
		expect(screen.getByRole('listbox')).toBeInTheDocument();
		await user.keyboard('{ArrowDown}{Enter}');

		expect(onChange).toHaveBeenCalledWith('patent');
		expect(screen.queryByRole('listbox')).not.toBeInTheDocument();

		await user.keyboard('{Space}{End}');
		expect(screen.getByRole('combobox')).toHaveAttribute(
			'aria-activedescendant',
			screen.getByRole('option', { name: 'ОСНО' }).id,
		);
		await user.keyboard('{Escape}');
		expect(screen.queryByRole('listbox')).not.toBeInTheDocument();
		expect(onChange).toHaveBeenCalledTimes(1);
	});

	it('Home возвращает к первому варианту, первая буква переходит к подходящему', async () => {
		const user = userEvent.setup();
		const onChange = vi.fn();
		render(<Controlled onChange={onChange} />);

		screen.getByRole('combobox').focus();
		await user.keyboard('о');
		expect(onChange).toHaveBeenLastCalledWith('osno');

		await user.keyboard('{Enter}{Home}');
		expect(screen.getByRole('combobox')).toHaveAttribute('aria-activedescendant', screen.getByRole('option', { name: 'УСН' }).id);
	});

	it('закрывается кликом вне списка', async () => {
		const user = userEvent.setup();
		render(<><Controlled /><button type="button">Снаружи</button></>);

		await user.click(screen.getByRole('combobox'));
		await user.click(screen.getByRole('button', { name: 'Снаружи' }));

		expect(screen.queryByRole('listbox')).not.toBeInTheDocument();
	});

	it('в заблокированном состоянии не открывается', async () => {
		const user = userEvent.setup();
		render(<Select aria-label="Режим" options={OPTIONS} value="usn" onChange={vi.fn()} disabled />);

		await user.click(screen.getByRole('combobox'));

		expect(screen.queryByRole('listbox')).not.toBeInTheDocument();
	});

	it('добавляет скрытое поле, если задано имя', () => {
		const { container } = render(<Select aria-label="Режим" name="tax" options={OPTIONS} value="patent" onChange={vi.fn()} />);

		expect(container.querySelector('input[type="hidden"][name="tax"]')).toHaveValue('patent');
	});
});

describe('SelectField', () => {
	it('связывает подпись, подсказку и ошибку с выпадающим списком', () => {
		render(<SelectField label="Режим" hint="Подсказка" error="Выберите режим" options={OPTIONS} value="" onChange={vi.fn()} />);

		const trigger = screen.getByRole('combobox', { name: 'Режим' });
		expect(trigger).toHaveAccessibleDescription('Подсказка Выберите режим');
		expect(trigger).toHaveAttribute('aria-invalid', 'true');
	});
});
