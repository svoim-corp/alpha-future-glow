const sizeFormatter = new Intl.NumberFormat('ru-RU', { maximumFractionDigits: 1 });

export const formatFileSize = (bytes: number): string => (
	bytes >= 1024 * 1024
		? `${sizeFormatter.format(bytes / 1024 / 1024)} МБ`
		: `${sizeFormatter.format(Math.max(1, Math.round(bytes / 1024)))} КБ`
);

export const usageLabel = (usage: { applications: number; label: string | null }): string => {
	if (usage.label) {
		return usage.label;
	}

	if (usage.applications === 0) {
		return 'Ещё не использован';
	}

	return usage.applications === 1 ? 'Использован в 1 заявке' : `Использован в ${usage.applications} заявках`;
};
