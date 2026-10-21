import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { Tooltip } from './Tooltip';

describe('Tooltip', () => {
	it('показывает пояснение при наведении и связывает с кнопкой', async () => {
		const user = userEvent.setup();
		render(<Tooltip label="Помощь" content="Краткое пояснение" />);
		expect(screen.queryByRole('tooltip')).not.toBeInTheDocument();
		await user.hover(screen.getByRole('button', { name: 'Помощь' }));
		expect(screen.getByRole('tooltip')).toHaveTextContent('Краткое пояснение');
		expect(screen.getByRole('button')).toHaveAccessibleDescription('Краткое пояснение');
		await user.unhover(screen.getByRole('button'));
		await waitFor(() => expect(screen.queryByRole('tooltip')).not.toBeInTheDocument());
	});

	it('открывается с клавиатуры, закрывается Escape и при потере фокуса', async () => {
		const user = userEvent.setup();
		render(<><Tooltip label="Помощь" content="Пояснение" /><button type="button">Дальше</button></>);
		await user.tab();
		expect(screen.getByRole('tooltip')).toBeInTheDocument();
		await user.keyboard('{Escape}');
		expect(screen.queryByRole('tooltip')).not.toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'Помощь' })).toHaveFocus();
		await user.tab();
		await user.tab({ shift: true });
		expect(screen.getByRole('tooltip')).toBeInTheDocument();
		await user.tab();
		expect(screen.queryByRole('tooltip')).not.toBeInTheDocument();
	});

	it('открывается касанием и закрывается кликом снаружи', async () => {
		const user = userEvent.setup();
		render(<><Tooltip label="Помощь" content="Пояснение" /><button type="button">Снаружи</button></>);
		await user.click(screen.getByRole('button', { name: 'Помощь' }));
		expect(screen.getByRole('tooltip')).toBeInTheDocument();
		await user.click(screen.getByRole('button', { name: 'Снаружи' }));
		expect(screen.queryByRole('tooltip')).not.toBeInTheDocument();
	});

	it('остаётся открытым при перемещении указателя на текст подсказки', async () => {
		const user = userEvent.setup();
		render(<Tooltip label="Помощь" content="Пояснение" />);
		await user.hover(screen.getByRole('button'));
		await user.hover(screen.getByRole('tooltip'));
		expect(screen.getByRole('tooltip')).toBeInTheDocument();
		await user.unhover(screen.getByRole('tooltip'));
		await waitFor(() => expect(screen.queryByRole('tooltip')).not.toBeInTheDocument());
	});
});
