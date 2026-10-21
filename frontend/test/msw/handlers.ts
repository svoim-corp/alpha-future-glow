import { http, type HttpHandler,HttpResponse } from 'msw';

import { apiUrl } from './apiUrl';

/**
 * Обработчики «счастливого пути» для всех тестов.
 * В конкретном тесте переопределяй через server.use(...).
 */
export const handlers: HttpHandler[] = [
	// колокольчик в шапке запрашивает уведомления на любой клиентской странице
	http.get(apiUrl('/notifications'), () => HttpResponse.json({ unreadCount: 0, items: [] })),
	// главная запрашивает кредиты, документы и подбор по планам на любой клиентской странице
	http.get(apiUrl('/loans'), () => HttpResponse.json([])),
	http.get(apiUrl('/plans/recommendations'), () => HttpResponse.json([])),
	http.get(apiUrl('/documents'), () => HttpResponse.json({
		items: [],
		counts: { ALL: 0, COMPANY: 0, STATEMENTS: 0, CONTRACTS: 0, PHOTOS: 0 },
	})),
];
