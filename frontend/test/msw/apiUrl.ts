import { env } from '@shared/config';

/** Полный URL эндпоинта так, как его запрашивает RTK Query в тестах */
export const apiUrl = (path: string) => `${env.apiBaseUrl}${path}`;
