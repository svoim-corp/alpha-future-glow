import { HomePage } from '@pages/home';
import { NotFoundPage } from '@pages/not-found';
import { ROUTES } from '@shared/config';
import { createBrowserRouter, type RouteObject } from 'react-router';

export const routes: RouteObject[] = [
	{ path: ROUTES.home, element: <HomePage /> },
	{ path: '*', element: <NotFoundPage /> },
];

export const createAppRouter = () => createBrowserRouter(routes);
