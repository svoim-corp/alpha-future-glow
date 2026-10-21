import { setupServer } from 'msw/node';

import { handlers } from './handlers';

export const server = setupServer(...handlers);
export { apiUrl } from './apiUrl';
