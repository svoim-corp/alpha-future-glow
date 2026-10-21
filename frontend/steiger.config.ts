import fsd from '@feature-sliced/steiger-plugin';
import { defineConfig } from 'steiger';

export default defineConfig([
	...fsd.configs.recommended,
	{
		// На старте проекта почти каждый слайс используется в одном месте —
		// правило шумит. Решение «слить слайс или оставить» принимаем на ревью.
		rules: {
			'fsd/insignificant-slice': 'off',
		},
	},
]);
