import { draftApplication, infoRequestedApplication } from '@test/fixtures/applications';
import { renderWithProviders } from '@test/render';
import { screen } from '@testing-library/react';

import { ApplicationCard } from './ApplicationCard';

describe('ApplicationCard', () => {
	it('для заявки с запросом банка показывает статус, бейдж «Нужно действие» и ссылку на карточку', () => {
		renderWithProviders(<ApplicationCard application={infoRequestedApplication} />);

		expect(screen.getByRole('heading', { name: 'Оборотный кредит' })).toBeInTheDocument();
		expect(screen.getByText('КЗ-2026-001024')).toBeInTheDocument();
		expect(screen.getByText('Нужна информация')).toBeInTheDocument();
		expect(screen.getByText('Нужно действие')).toBeInTheDocument();
		expect(screen.getByRole('link', { name: 'Открыть заявку' })).toHaveAttribute('href', '/applications/a-1024');
	});

	it('черновик без номера ведёт в мастер и принимает дополнительные действия', () => {
		renderWithProviders(<ApplicationCard application={draftApplication} actions={<button type="button">Удалить</button>} />);

		expect(screen.getByText('Номер появится после отправки')).toBeInTheDocument();
		expect(screen.getByRole('link', { name: 'Продолжить заполнение' })).toHaveAttribute('href', '/applications/a-1031/edit');
		expect(screen.getByRole('button', { name: 'Удалить' })).toBeInTheDocument();
		expect(screen.queryByText('Нужно действие')).not.toBeInTheDocument();
	});
});
