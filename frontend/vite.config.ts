import { fileURLToPath, URL } from 'node:url';

import react from '@vitejs/plugin-react';
import { defineConfig, loadEnv } from 'vite';
import svgr from 'vite-plugin-svgr';

const fromRoot = (path: string) => fileURLToPath(new URL(path, import.meta.url));

export const alias = {
	'@app': fromRoot('./src/app'),
	'@pages': fromRoot('./src/pages'),
	'@widgets': fromRoot('./src/widgets'),
	'@features': fromRoot('./src/features'),
	'@entities': fromRoot('./src/entities'),
	'@shared': fromRoot('./src/shared'),
	'@test': fromRoot('./test'),
};

export default defineConfig(({ mode }) => {
	const env = loadEnv(mode, process.cwd(), '');

	return {
		plugins: [react(), svgr()],
		resolve: { alias },
		css: {
			modules: { localsConvention: 'camelCaseOnly' },
		},
		server: {
			port: 3000,
			proxy: {
				'/api': {
					target: env.VITE_API_PROXY_TARGET ?? 'http://localhost:8080',
					changeOrigin: true,
				},
			},
		},
		preview: { port: 3000 },
	};
});
