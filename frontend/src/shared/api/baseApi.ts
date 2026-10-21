import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';

import { env } from '../config';

/**
 * Единый RTK Query API. Эндпоинты добавляются в слайсах entities/features
 * через `baseApi.injectEndpoints`, а не здесь.
 */
export const baseApi = createApi({
	reducerPath: 'api',
	baseQuery: fetchBaseQuery({
		baseUrl: env.apiBaseUrl,
		credentials: 'include',
	}),
	tagTypes: [
		'Session',
		'Application',
		'Product',
		'Loan',
		'Notification',
		'Ticket',
		'BankApplication',
		'Plan',
	],
	endpoints: () => ({}),
});
