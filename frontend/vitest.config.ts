import { defineConfig, mergeConfig } from 'vitest/config';

import viteConfig from './vite.config.ts';

export default defineConfig((env) => mergeConfig(viteConfig(env), defineConfig({
	test: {
		environment: 'jsdom',
		globals: true,
		setupFiles: ['./test/setup.ts'],
		include: ['src/**/*.test.{ts,tsx}', 'test/**/*.test.{ts,tsx}'],
		css: { modules: { classNameStrategy: 'non-scoped' } },
		env: {
			VITE_API_BASE_URL: 'http://localhost/api/v1',
		},
		restoreMocks: true,
		coverage: {
			provider: 'v8',
			reporter: ['text', 'html', 'lcov'],
			include: ['src/**/*.{ts,tsx}'],
			exclude: [
				'src/**/index.ts',
				'src/**/*.d.ts',
				'src/**/*.test.{ts,tsx}',
				'src/app/entrypoint/main.tsx',
				// моки для демо без бэкенда — инструмент разработки, не код продукта
				'src/app/mocks/**',
			],
			thresholds: {
				lines: 80,
				functions: 80,
				statements: 80,
				branches: 70,
			},
		},
	},
})));
