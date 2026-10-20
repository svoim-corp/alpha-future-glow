import { ApplicationPage } from '@pages/application';
import { ApplicationEditPage } from '@pages/application-edit';
import { ApplicationSignPage } from '@pages/application-sign';
import { CreditsPage } from '@pages/credits';
import { CreditsCalendarPage } from '@pages/credits-calendar';
import { HomePage } from '@pages/home';
import { LoginPage } from '@pages/login';
import { NotFoundPage } from '@pages/not-found';
import { ProductsPage } from '@pages/products';
import { SupportPage } from '@pages/support';
import { ROUTES } from '@shared/config';
import { createBrowserRouter, type RouteObject } from 'react-router';

import { BankLayout } from './BankLayout';
import { ClientLayout } from './ClientLayout';
import { RequireRole } from './RequireRole';

export const routes: RouteObject[] = [
	{ path: ROUTES.login, element: <LoginPage /> },
	{
		element: <RequireRole allow="CLIENT" />,
		children: [
			{
				element: <ClientLayout />,
				children: [
					{ path: ROUTES.home, element: <HomePage /> },
					{ path: ROUTES.products, element: <ProductsPage /> },
					{ path: ROUTES.applicationEdit, element: <ApplicationEditPage /> },
					{ path: ROUTES.application, element: <ApplicationPage /> },
					{ path: ROUTES.applicationSign, element: <ApplicationSignPage /> },
					{ path: ROUTES.credits, element: <CreditsPage /> },
					{ path: ROUTES.creditsCalendar, element: <CreditsCalendarPage /> },
					{ path: ROUTES.support, element: <SupportPage /> },
				],
			},
		],
	},
	{
		element: <RequireRole allow="BANK_OFFICER" />,
		children: [
			{
				element: <BankLayout />,
				children: [],
			},
		],
	},
	{ path: '*', element: <NotFoundPage /> },
];

export const createAppRouter = () => createBrowserRouter(routes);
