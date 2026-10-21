import { CATEGORIES } from '@entities/document';
import { getErrorMessage } from '@shared/api';
import { Alert, Button, FileUpload, Modal, SelectField } from '@shared/ui';
import { useState } from 'react';

import { useUploadDocumentMutation } from '../api/uploadApi';

const MAX_BYTES = 25 * 1024 * 1024;
const OPTIONS = CATEGORIES.filter((item) => item.value !== 'ALL').map(({ value, label }) => ({ value, label }));

export const UploadDocument = () => {
	const [upload, { isLoading }] = useUploadDocumentMutation();
	const [open, setOpen] = useState(false);
	const [file, setFile] = useState<File | null>(null);
	const [category, setCategory] = useState('COMPANY');
	const [fileError, setFileError] = useState<string | undefined>();
	const [failure, setFailure] = useState<string | null>(null);

	const close = () => {
		setOpen(false);
		setFile(null);
		setFileError(undefined);
		setFailure(null);
	};

	const submit = async () => {
		if (!file) {
			return;
		}

		setFailure(null);

		try {
			await upload({ file, category }).unwrap();
			close();
		} catch (error) {
			setFailure(getErrorMessage(error, 'Не удалось загрузить файл. Повторите попытку.'));
		}
	};

	return (
		<>
			<Button onClick={() => setOpen(true)}>Добавить документ</Button>
			{open && (
				<Modal
					title="Добавить документ"
					onClose={close}
					footer={(
						<>
							<Button variant="ghost" onClick={close}>Отмена</Button>
							<Button onClick={() => void submit()} loading={isLoading} disabled={!file}>
								Загрузить
							</Button>
						</>
					)}
				>
					<FileUpload
						label="Файл"
						accept=".pdf,.txt,.jpg,.jpeg,.png"
						hint="PDF, TXT, JPG или PNG, до 25 МБ"
						maxBytes={MAX_BYTES}
						file={file}
						error={fileError}
						onSelect={(next) => {
							setFile(next);
							setFileError(undefined);
						}}
						onReject={setFileError}
					/>
					<SelectField
						label="Категория"
						value={category}
						options={OPTIONS}
						onChange={(event) => setCategory(event.target.value)}
					/>
					{failure && <Alert tone="danger" title={failure} />}
				</Modal>
			)}
		</>
	);
};
