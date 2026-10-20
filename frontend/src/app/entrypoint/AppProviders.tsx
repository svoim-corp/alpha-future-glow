import { type ReactNode } from 'react';
import { Provider } from 'react-redux';

import { type AppStore } from '../model';

interface AppProvidersProps {
	store: AppStore;
	children: ReactNode;
}

export const AppProviders = ({ store, children }: AppProvidersProps) => (
	<Provider store={store}>
		{children}
	</Provider>
);
