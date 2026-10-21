import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { FileUpload } from './FileUpload';

const setup = (extra: { file?: File | null; error?: string } = {}) => {
	const onSelect = vi.fn();
	const onReject = vi.fn();
	render(
		<FileUpload
			label="Выписка из 1С"
			accept=".txt"
			hint="Файл .txt до 10 МБ"
			maxBytes={1024}
			file={extra.file ?? null}
			error={extra.error}
			onSelect={onSelect}
			onReject={onReject}
		/>,
	);

	return { onSelect, onReject };
};

describe('FileUpload', () => {
	it('передаёт выбранный файл', async () => {
		const { onSelect } = setup();
		const file = new File(['abc'], 'statement.txt', { type: 'text/plain' });

		await userEvent.upload(screen.getByLabelText('Выписка из 1С'), file);

		expect(onSelect).toHaveBeenCalledWith(file);
	});

	it('отклоняет файл больше лимита и не выбирает его', async () => {
		const { onSelect, onReject } = setup();
		const file = new File([new Uint8Array(2048)], 'big.txt', { type: 'text/plain' });

		await userEvent.upload(screen.getByLabelText('Выписка из 1С'), file);

		expect(onReject).toHaveBeenCalledOnce();
		expect(onSelect).not.toHaveBeenCalled();
	});

	it('показывает имя файла и ошибку рядом с ним', () => {
		setup({ file: new File(['a'], 'statement.txt'), error: 'Не формат 1С' });

		expect(screen.getByText(/statement\.txt/)).toBeInTheDocument();
		expect(screen.getByRole('alert')).toHaveTextContent('Не формат 1С');
	});
});
