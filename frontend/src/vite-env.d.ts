/// <reference types="vite/client" />

interface ImportMetaEnv {
	readonly VITE_API_BASE_URL?: string;
	/** 'true' — запросы к API обслуживают моки в браузере (MSW), бэкенд не нужен */
	readonly VITE_MOCK?: string;
}

interface ImportMeta {
	readonly env: ImportMetaEnv;
}
