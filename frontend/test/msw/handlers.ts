import { http, type HttpHandler,HttpResponse } from 'msw';

import { apiUrl } from './apiUrl';

/**
 * Обработчики «счастливого пути» для всех тестов.
 * В конкретном тесте переопределяй через server.use(...).
 */
export const handlers: HttpHandler[] = [
	// колокольчик в шапке запрашивает уведомления на любой клиентской странице
	http.get(apiUrl('/notifications'), () => HttpResponse.json({ unreadCount: 0, items: [] })),
];
