import { chooseOption } from '@test/select';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { useState } from 'react';

import { MonthPicker } from './MonthPicker';

const CalendarMonth = () => {
	const [value, setValue] = useState({ year: 2026, month: 9 });

	return <MonthPicker value={value} onChange={setValue} />;
};

describe('MonthPicker', () => {
	it('выбирает месяц и год независимо', async () => {
		const user = userEvent.setup();
		render(<CalendarMonth />);
		await chooseOption(user, 'Месяц', 'февраль');
		expect(screen.getByLabelText('Год')).toHaveTextContent('2026');
		await chooseOption(user, 'Год', '2028');
		expect(screen.getByLabelText('Месяц')).toHaveTextContent('февраль');
		expect(screen.getByLabelText('Год')).toHaveTextContent('2028');
	});

	it('сохраняет возможность выбрать год отображаемого плана за пределами базового списка', () => {
		render(<MonthPicker value={{ year: 2150, month: 0 }} onChange={vi.fn()} />);
		expect(screen.getByLabelText('Год')).toHaveTextContent('2150');
	});
});
