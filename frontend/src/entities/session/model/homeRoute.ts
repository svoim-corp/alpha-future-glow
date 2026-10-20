import { ROUTES } from '@shared/config';

import { type UserRole } from './types';

/** Стартовая страница роли: клиент — главная, сотрудник — очередь заявок */
export const homeRouteFor = (role: UserRole): string => (role === 'BANK_OFFICER' ? ROUTES.bankQueue : ROUTES.home);
