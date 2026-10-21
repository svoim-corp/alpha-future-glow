import { useCallback, useEffect, useState } from 'react';

type Theme = 'light' | 'dark';

const STORAGE_KEY = 'credit-ui-theme';

const readSaved = (): Theme | null => {
	try {
		const saved = localStorage.getItem(STORAGE_KEY);

		return saved === 'light' || saved === 'dark' ? saved : null;
	} catch {
		return null;
	}
};

const systemTheme = (): Theme => (globalThis.matchMedia?.('(prefers-color-scheme: dark)').matches ? 'dark' : 'light');

/** Применяет сохранённую или системную тему к странице до отрисовки: так вход и 404 тоже следуют теме */
export const applyTheme = () => {
	document.documentElement.dataset.theme = readSaved() ?? systemTheme();
};

/** Тема оформления: выбор пользователя запоминается, иначе берётся тема системы */
export const useTheme = () => {
	const [theme, setTheme] = useState<Theme>(() => readSaved() ?? systemTheme());

	useEffect(() => {
		document.documentElement.dataset.theme = theme;
	}, [theme]);

	const toggle = useCallback(() => {
		const next: Theme = theme === 'dark' ? 'light' : 'dark';
		setTheme(next);

		try {
			localStorage.setItem(STORAGE_KEY, next);
		} catch {
			// хранилище недоступно — тема действует только в этой сессии
		}
	}, [theme]);

	return { theme, toggle };
};
