/** Шаблоны маршрутов для роутера */
export const ROUTES = {
	home: '/',
	login: '/login',
	products: '/products',
	applicationEdit: '/applications/:id/edit',
	application: '/applications/:id',
	applicationSign: '/applications/:id/sign',
	credits: '/credits',
	creditsCalendar: '/credits/calendar',
	support: '/support',
	bankQueue: '/bank/queue',
	bankApplication: '/bank/applications/:id',
} as const;

/** Построители ссылок с подстановкой id */
export const paths = {
	applicationEdit: (id: string) => `/applications/${id}/edit`,
	application: (id: string) => `/applications/${id}`,
	applicationSign: (id: string) => `/applications/${id}/sign`,
	bankApplication: (id: string) => `/bank/applications/${id}`,
	support: (applicationId?: string) => (applicationId ? `/support?applicationId=${applicationId}` : '/support'),
} as const;
