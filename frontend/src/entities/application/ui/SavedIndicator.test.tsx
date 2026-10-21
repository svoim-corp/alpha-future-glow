import { render, screen } from '@testing-library/react';

import { SavedIndicator } from './SavedIndicator';

describe('SavedIndicator', () => {
	it('до первого сохранения сообщает об этом', () => {
		render(<SavedIndicator savedAt={null} saving={false} failed={false} />);

		expect(screen.getByRole('status')).toHaveTextContent('Черновик ещё не сохранялся');
	});

	it('показывает время сохранения, идущее сохранение и сбой', () => {
		const { rerender } = render(<SavedIndicator savedAt="2026-10-20T10:12:00Z" saving={false} failed={false} />);
		expect(screen.getByRole('status')).toHaveTextContent(/Черновик сохранён 20\.10\.2026/);

		rerender(<SavedIndicator savedAt="2026-10-20T10:12:00Z" saving failed={false} />);
		expect(screen.getByRole('status')).toHaveTextContent('Сохраняем…');

		rerender(<SavedIndicator savedAt="2026-10-20T10:12:00Z" saving={false} failed />);
		expect(screen.getByRole('status')).toHaveTextContent('Не удалось сохранить черновик');
	});
});
