import { type ChangeEvent, useId } from 'react';

import styles from './FileUpload.module.scss';

export interface FileUploadProps {
	label: string;
	accept: string;
	hint: string;
	/** Лимит размера в байтах — проверяется до отправки */
	maxBytes: number;
	file: File | null;
	error?: string;
	onSelect: (file: File | null) => void;
	onReject: (reason: string) => void;
	disabled?: boolean;
}

export const FileUpload = ({
	label,
	accept,
	hint,
	maxBytes,
	file,
	error,
	onSelect,
	onReject,
	disabled,
}: FileUploadProps) => {
	const id = useId();
	const hintId = `${id}-hint`;
	const errorId = `${id}-error`;

	const handleChange = (event: ChangeEvent<HTMLInputElement>) => {
		const picked = event.target.files?.[0] ?? null;

		if (picked && picked.size > maxBytes) {
			onReject(`Файл больше ${Math.round(maxBytes / (1024 * 1024))} МБ`);
			event.target.value = '';

			return;
		}

		onSelect(picked);
	};

	return (
		<div className={styles.upload}>
			<label className={styles.label} htmlFor={id}>{label}</label>
			<input
				id={id}
				type="file"
				accept={accept}
				disabled={disabled}
				onChange={handleChange}
				aria-describedby={error ? `${hintId} ${errorId}` : hintId}
				aria-invalid={error ? true : undefined}
			/>
			<p className={styles.hint} id={hintId}>{hint}</p>
			{file && <div className={styles.name}>Выбран файл: {file.name}</div>}
			{error && <p className={styles.error} id={errorId} role="alert">{error}</p>}
		</div>
	);
};
