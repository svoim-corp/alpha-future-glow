import { formatDateTime } from '@shared/lib/format';

interface SavedIndicatorProps {
	savedAt: string | null;
	saving: boolean;
	failed: boolean;
}

/** «Черновик сохранён» показываем только после подтверждения сервера */
export const SavedIndicator = ({ savedAt, saving, failed }: SavedIndicatorProps) => {
	let text = 'Черновик ещё не сохранялся';

	if (saving) {
		text = 'Сохраняем…';
	} else if (failed) {
		text = 'Не удалось сохранить черновик';
	} else if (savedAt) {
		text = `Черновик сохранён ${formatDateTime(savedAt)}`;
	}

	return <p role="status" aria-live="polite">{text}</p>;
};
