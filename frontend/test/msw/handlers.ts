import { type HttpHandler } from 'msw';

/**
 * Обработчики «счастливого пути» для всех тестов.
 * Пример: http.get(apiUrl('/applications'), () => HttpResponse.json(fixture)).
 * В конкретном тесте переопределяй через server.use(...).
 */
export const handlers: HttpHandler[] = [];
