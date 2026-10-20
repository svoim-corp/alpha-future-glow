import { baseApi } from '@shared/api';

import { type PlanHomeRecommendation, type PlansResponse } from '../model/types';

export const planApi = baseApi.injectEndpoints({
	endpoints: (build) => ({
		getPlans: build.query<PlansResponse, void>({
			query: () => '/plans',
			providesTags: [{ type: 'Plan', id: 'LIST' }],
		}),
		getPlanRecommendations: build.query<PlanHomeRecommendation[], void>({
			query: () => '/plans/recommendations',
			providesTags: [{ type: 'Plan', id: 'LIST' }],
		}),
	}),
});

export const { useGetPlansQuery, useGetPlanRecommendationsQuery } = planApi;
