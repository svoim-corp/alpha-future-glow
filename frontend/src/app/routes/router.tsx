import { ApplicationPage } from '@pages/application';
import { ApplicationEditPage } from '@pages/application-edit';
import { ApplicationSignPage } from '@pages/application-sign';
import { ApplicationsPage } from '@pages/applications';
import { BankApplicationPage } from '@pages/bank-application';
import { BankQueuePage } from '@pages/bank-queue';
import { CreditsPage } from '@pages/credits';
import { CreditsCalendarPage } from '@pages/credits-calendar';
import { DocumentsPage } from '@pages/documents';
import { HomePage } from '@pages/home';
import { LoginPage } from '@pages/login';
import { NotFoundPage } from '@pages/not-found';
import { PlannerPage } from '@pages/planner';
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
					{ path: ROUTES.documents, element: <DocumentsPage /> },
					{ path: ROUTES.planner, element: <PlannerPage /> },
					{ path: ROUTES.applications, element: <ApplicationsPage /> },
				],
			},
		],
	},
	{
		element: <RequireRole allow="BANK_OFFICER" />,
		children: [
			{
				element: <BankLayout />,
				children: [
					{ path: ROUTES.bankQueue, element: <BankQueuePage /> },
					{ path: ROUTES.bankApplication, element: <BankApplicationPage /> },
				],
			},
		],
	},
	{ path: '*', element: <NotFoundPage /> },
];

export const createAppRouter = () => createBrowserRouter(routes);
