import { dashboardCustomizationSlice } from '@features/customize-dashboard';
import { combineSlices, configureStore } from '@reduxjs/toolkit';
import { setupListeners } from '@reduxjs/toolkit/query';
import { baseApi } from '@shared/api';

/**
 * Корневой редьюсер. Новый слайс (entities/features) регистрируется здесь —
 * импортом из public API слайса: combineSlices(baseApi, someSlice, ...).
 */
export const rootReducer = combineSlices(baseApi, dashboardCustomizationSlice);

export type AppRootState = ReturnType<typeof rootReducer>;

export const setupStore = (preloadedState?: Partial<AppRootState>) => {
	const store = configureStore({
		reducer: rootReducer,
		preloadedState,
		middleware: (getDefaultMiddleware) => getDefaultMiddleware().concat(baseApi.middleware),
	});
	setupListeners(store.dispatch);

	return store;
};

export type AppStore = ReturnType<typeof setupStore>;
