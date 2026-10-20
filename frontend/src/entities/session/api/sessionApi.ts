import { baseApi } from '@shared/api';

import { type CurrentUser, type LoginRequest, type LoginResponse } from '../model/types';

export const sessionApi = baseApi.injectEndpoints({
	endpoints: (build) => ({
		getMe: build.query<CurrentUser, void>({
			query: () => '/auth/me',
			providesTags: ['Session'],
		}),
		login: build.mutation<LoginResponse, LoginRequest>({
			query: (body) => ({ url: '/auth/login', method: 'POST', body }),
			invalidatesTags: ['Session'],
		}),
		logout: build.mutation<void, void>({
			query: () => ({ url: '/auth/logout', method: 'POST' }),
			// после выхода сбрасываем весь серверный кэш, чтобы данные не утекли в следующую сессию
			onQueryStarted: (_arg, { dispatch, queryFulfilled }) => {
				queryFulfilled.then(() => dispatch(baseApi.util.resetApiState())).catch(() => undefined);
			},
		}),
	}),
});

export const { useGetMeQuery, useLoginMutation, useLogoutMutation } = sessionApi;
