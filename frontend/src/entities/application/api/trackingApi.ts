import { baseApi } from '@shared/api';

import { type InfoRequest, type Offer, type Timeline } from '../model/tracking';

export const trackingApi = baseApi.injectEndpoints({
	endpoints: (build) => ({
		getTimeline: build.query<Timeline, string>({
			query: (id) => `/applications/${id}/timeline`,
			providesTags: (_result, _error, id) => [{ type: 'Application', id }],
		}),
		getInfoRequests: build.query<InfoRequest[], string>({
			query: (id) => `/applications/${id}/info-requests`,
			providesTags: (_result, _error, id) => [{ type: 'Application', id }],
		}),
		getOffer: build.query<Offer, string>({
			query: (id) => `/applications/${id}/offer`,
			providesTags: (_result, _error, id) => [{ type: 'Application', id }],
		}),
	}),
});

export const { useGetTimelineQuery, useGetInfoRequestsQuery, useGetOfferQuery } = trackingApi;
