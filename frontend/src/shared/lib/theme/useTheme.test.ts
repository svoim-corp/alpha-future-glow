import { applyTheme } from './useTheme';

describe('applyTheme', () => {
	afterEach(() => {
		localStorage.removeItem('credit-ui-theme');
		delete document.documentElement.dataset.theme;
	});

	it('применяет сохранённую тему к странице', () => {
		localStorage.setItem('credit-ui-theme', 'dark');

		applyTheme();

		expect(document.documentElement.dataset.theme).toBe('dark');
	});

	it('без сохранённого выбора берёт тему системы', () => {
		applyTheme();

		expect(document.documentElement.dataset.theme).toMatch(/^(light|dark)$/u);
	});
});
