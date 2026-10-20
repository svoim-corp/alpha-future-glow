import { type AppRootState, type AppStore } from './store';

/**
 * Глобальные типы стора — чтобы shared/lib/store мог типизировать хуки,
 * не импортируя слой app (FSD: shared не знает о вышележащих слоях).
 */
declare global {
	type RootState = AppRootState;
	type AppDispatch = AppStore['dispatch'];
}

export {};
