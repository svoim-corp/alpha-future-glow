import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { useState } from 'react';

import { MoneyField } from './MoneyField';

const MoneyInput = ({ initial = '' }: { initial?: string }) => {
	const [value, setValue] = useState(initial);

	return <><MoneyField label="Бюджет" value={value} onValueChange={setValue} /><output aria-label="Рубли">{value}</output></>;
};

describe('MoneyField', () => {
	it('показывает разделители тысяч, но хранит только цифры', async () => {
		const user = userEvent.setup();
		render(<MoneyInput />);
		await user.type(screen.getByLabelText('Бюджет'), '1200000');
		expect(screen.getByLabelText('Бюджет')).toHaveValue('1 200 000');
		expect(screen.getByLabelText('Рубли')).toHaveTextContent('1200000');
		await user.clear(screen.getByLabelText('Бюджет'));
		expect(screen.getByLabelText('Рубли')).toBeEmptyDOMElement();
	});

	it('принимает вставку с пробелами, не превращает дробь и минус в другую сумму', async () => {
		const user = userEvent.setup();
		render(<MoneyInput />);
		await user.click(screen.getByLabelText('Бюджет'));
		await user.paste('12\u00a0345\u00a0678');
		expect(screen.getByLabelText('Бюджет')).toHaveValue('12 345 678');
		await user.clear(screen.getByLabelText('Бюджет'));
		await user.paste('-100');
		expect(screen.getByLabelText('Бюджет')).toHaveValue('');
		await user.paste('12,5');
		expect(screen.getByLabelText('Бюджет')).toHaveValue('');
	});

	it('позволяет редактировать середину суммы и удалять цифры рядом с разделителем', async () => {
		const user = userEvent.setup();
		render(<MoneyInput initial="1234567" />);
		const input = screen.getByRole<HTMLInputElement>('textbox');
		await user.click(input);
		await user.keyboard('{Home}{ArrowRight}{ArrowRight}{Backspace}');
		expect(input).toHaveValue('234 567');
		expect(input.selectionStart).toBe(0);
		await user.keyboard('{ArrowRight}{ArrowRight}{ArrowRight}{Delete}');
		expect(input).toHaveValue('23 467');
	});

	it('не округляет длинную строку при отображении и сохраняет disabled', async () => {
		const user = userEvent.setup();
		render(<MoneyField label="Бюджет" value="9007199254740993" onValueChange={vi.fn()} disabled />);
		expect(screen.getByLabelText('Бюджет')).toHaveValue('9 007 199 254 740 993');
		await user.type(screen.getByLabelText('Бюджет'), '1');
		expect(screen.getByLabelText('Бюджет')).toBeDisabled();
	});
});
