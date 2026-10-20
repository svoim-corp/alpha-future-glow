import { baseApi } from '@shared/api';

import { type NotificationsResponse } from '../model/types';

export const notificationApi = baseApi.injectEndpoints({
	endpoints: (build) => ({
		getNotifications: build.query<NotificationsResponse, void>({
			query: () => '/notifications',
			providesTags: [{ type: 'Notification', id: 'LIST' }],
		}),
		markNotificationRead: build.mutation<void, string>({
			query: (id) => ({ url: `/notifications/${id}/read`, method: 'POST' }),
			invalidatesTags: [{ type: 'Notification', id: 'LIST' }],
		}),
	}),
});

export const { useGetNotificationsQuery, useMarkNotificationReadMutation } = notificationApi;
