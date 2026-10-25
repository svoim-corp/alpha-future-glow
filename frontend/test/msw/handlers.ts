import { http, type HttpHandler,HttpResponse } from 'msw';

import { clientUser } from '../fixtures/session';
import { apiUrl } from './apiUrl';

/**
 * Обработчики «счастливого пути» для всех тестов.
 * В конкретном тесте переопределяй через server.use(...).
 */
export const handlers: HttpHandler[] = [
	http.get(apiUrl('/support/conversations'), () => HttpResponse.json([])),
	// заголовок страниц берёт название компании из текущего пользователя
	http.get(apiUrl('/auth/me'), () => HttpResponse.json(clientUser)),
	// колокольчик в шапке запрашивает уведомления на любой клиентской странице
	http.get(apiUrl('/notifications'), () => HttpResponse.json({ unreadCount: 0, items: [] })),
	// главная запрашивает кредиты и подбор по планам на любой клиентской странице
	http.get(apiUrl('/loans'), () => HttpResponse.json([])),
	http.get(apiUrl('/plans/recommendations'), () => HttpResponse.json([])),
	// страница заявки запрашивает запросы банка при любом статусе, кроме черновика
	http.get(apiUrl('/applications/:id/info-requests'), () => HttpResponse.json([])),
];
