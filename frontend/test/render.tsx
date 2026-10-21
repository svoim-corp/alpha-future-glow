import { AppProviders } from '@app/entrypoint';
import { type AppRootState, type AppStore, setupStore } from '@app/model';
import { render, type RenderOptions } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { type ReactElement, type ReactNode } from 'react';
import { MemoryRouter } from 'react-router';

interface ExtendedRenderOptions extends Omit<RenderOptions, 'wrapper'> {
	preloadedState?: Partial<AppRootState>;
	store?: AppStore;
	/** Начальный адрес для MemoryRouter */
	route?: string;
	/** false — если тест сам рендерит RouterProvider */
	withRouter?: boolean;
}

/**
 * Рендер с реальным стором (свежим на каждый тест), роутером и userEvent.
 * Сеть мокается через MSW (test/msw), а не через моки RTK Query.
 */
export const renderWithProviders = (
	ui: ReactElement,
	{
		preloadedState,
		store = setupStore(preloadedState),
		route = '/',
		withRouter = true,
		...options
	}: ExtendedRenderOptions = {},
) => {
	const Wrapper = ({ children }: { children: ReactNode }) => (
		<AppProviders store={store}>
			{withRouter ? <MemoryRouter initialEntries={[route]}>{children}</MemoryRouter> : children}
		</AppProviders>
	);

	return {
		store,
		user: userEvent.setup(),
		...render(ui, { wrapper: Wrapper, ...options }),
	};
};
