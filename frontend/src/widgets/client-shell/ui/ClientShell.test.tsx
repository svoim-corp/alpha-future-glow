import { clientUser } from '@test/fixtures/session';
import { apiUrl, server } from '@test/msw/server';
import { renderWithProviders } from '@test/render';
import { screen } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { ClientShell } from './ClientShell';

const shell = (route = '/') => {
	server.use(http.get(apiUrl('/auth/me'), () => HttpResponse.json(clientUser)));

	return renderWithProviders(<ClientShell><p>Содержимое</p></ClientShell>, { route });
};

describe('ClientShell', () => {
	it('показывает боковую навигацию клиента и содержимое страницы', () => {
		shell();

		const nav = screen.getByRole('navigation', { name: 'Разделы кабинета' });

		expect(nav).toHaveTextContent('Главная');
		expect(screen.getByRole('link', { name: 'Планер' })).toHaveAttribute('href', '/planner');
		expect(screen.getByRole('link', { name: 'Мои заявки' })).toHaveAttribute('href', '/applications');
		expect(screen.getByRole('link', { name: 'Мои кредиты' })).toHaveAttribute('href', '/credits');
		expect(screen.getByRole('link', { name: 'Каталог' })).toHaveAttribute('href', '/products');
		expect(screen.getByRole('link', { name: 'Документы' })).toHaveAttribute('href', '/documents');
		expect(screen.getByRole('link', { name: 'Поддержка' })).toHaveAttribute('href', '/support');
		expect(screen.getByRole('button', { name: 'Выйти' })).toBeInTheDocument();
		expect(screen.getByText('Содержимое')).toBeInTheDocument();
	});

	it('текущий раздел помечен aria-current', () => {
		shell('/planner');

		expect(screen.getByRole('link', { name: 'Планер' })).toHaveAttribute('aria-current', 'page');
		expect(screen.getByRole('link', { name: 'Главная' })).not.toHaveAttribute('aria-current');
	});

	it('профиль раскрывается и показывает должность', async () => {
		const { user } = shell();

		await user.click(await screen.findByRole('button', { name: /Смирнов Олег Викторович/ }));

		expect(screen.getByText(clientUser.position)).toBeInTheDocument();
	});

	it('тема переключается и запоминается', async () => {
		const { user } = shell();

		await user.click(screen.getByRole('button', { name: /тема/ }));

		expect(document.documentElement.dataset.theme).toMatch(/light|dark/);
		expect(screen.getByRole('button', { name: /тема/ })).toHaveAttribute('aria-pressed');
	});

	it('мобильное меню открывается и закрывается по Escape', async () => {
		const { user } = shell();
		const menu = screen.getByRole('button', { name: 'Меню' });

		await user.click(menu);

		expect(menu).toHaveAttribute('aria-expanded', 'true');

		await user.keyboard('{Escape}');

		expect(menu).toHaveAttribute('aria-expanded', 'false');
	});
});
