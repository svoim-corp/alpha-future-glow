import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { useState } from 'react';

import { CheckboxField, Field, RadioGroup, TextAreaField } from './index';

describe('Field', () => {
	it('связывает подпись с полем', () => {
		render(<Field label="ИНН" />);

		expect(screen.getByLabelText('ИНН')).toBeInTheDocument();
	});

	it('показывает подсказку и ошибку, связывая их через aria-describedby', () => {
		render(<Field label="Сумма" hint="От 300 000 ₽" error="Сумма вне рамок" />);

		const input = screen.getByLabelText('Сумма');

		expect(input).toHaveAttribute('aria-invalid', 'true');
		expect(input).toHaveAccessibleDescription('От 300 000 ₽ Сумма вне рамок');
		expect(screen.getByRole('alert')).toHaveTextContent('Сумма вне рамок');
	});

	it('без ошибки не помечается невалидным', () => {
		render(<Field label="Сумма" />);

		expect(screen.getByLabelText('Сумма')).not.toHaveAttribute('aria-invalid');
	});
});

describe('TextAreaField', () => {
	it('принимает ввод и показывает ошибку', async () => {
		render(<TextAreaField label="Комментарий" error="Обязательное поле" />);

		await userEvent.type(screen.getByLabelText('Комментарий'), 'Привет');

		expect(screen.getByLabelText('Комментарий')).toHaveValue('Привет');
		expect(screen.getByRole('alert')).toHaveTextContent('Обязательное поле');
	});
});

describe('CheckboxField', () => {
	it('по умолчанию не отмечен и переключается кликом по подписи', async () => {
		render(<CheckboxField label="Согласен на запрос в БКИ" error="Нужно согласие" />);

		const checkbox = screen.getByRole('checkbox', { name: 'Согласен на запрос в БКИ' });
		expect(checkbox).not.toBeChecked();

		await userEvent.click(screen.getByText('Согласен на запрос в БКИ'));

		expect(checkbox).toBeChecked();
		expect(checkbox).toHaveAccessibleDescription('Нужно согласие');
	});
});

describe('RadioGroup', () => {
	const Controlled = () => {
		const [value, setValue] = useState<'a' | 'b' | null>(null);

		return (
			<RadioGroup
				legend="Как получить сумму"
				value={value}
				onChange={setValue}
				options={[{ value: 'a', label: 'Выписка' }, { value: 'b', label: 'Меньше' }]}
				error={value === null ? 'Выберите вариант' : undefined}
			/>
		);
	};

	it('группа подписана, выбор взаимоисключающий', async () => {
		render(<Controlled />);

		expect(screen.getByRole('group', { name: 'Как получить сумму' })).toBeInTheDocument();
		await userEvent.click(screen.getByRole('radio', { name: 'Меньше' }));

		expect(screen.getByRole('radio', { name: 'Меньше' })).toBeChecked();
		expect(screen.getByRole('radio', { name: 'Выписка' })).not.toBeChecked();
		expect(screen.queryByRole('alert')).not.toBeInTheDocument();
	});
});
