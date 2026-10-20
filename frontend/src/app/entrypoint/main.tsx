import '../styles/index.scss';

import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { RouterProvider } from 'react-router';

import { setupStore } from '../model';
import { createAppRouter } from '../routes';
import { AppProviders } from './AppProviders';

const container = document.getElementById('root');

if (!container) {
	throw new Error('Не найден #root в index.html');
}

const store = setupStore();
const router = createAppRouter();

createRoot(container).render(
	<StrictMode>
		<AppProviders store={store}>
			<RouterProvider router={router} />
		</AppProviders>
	</StrictMode>,
);
