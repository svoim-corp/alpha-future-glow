import { baseApi } from '@shared/api';

import {
	type BusinessPlan,
	type CreatePlanArgs,
	type PlanHomeRecommendation,
	type PlanPatch,
	type PlansResponse,
	type PlanState,
} from '../model/types';

export const planApi = baseApi.injectEndpoints({
	endpoints: (build) => ({
		getPlan: build.query<BusinessPlan, string>({
			query: (id) => `/plans/${encodeURIComponent(id)}`,
			providesTags: (_result, _error, id) => [{ type: 'Plan', id }, { type: 'Plan', id: 'LIST' }],
		}),
		createPlan: build.mutation<BusinessPlan, CreatePlanArgs>({
			query: (body) => ({ url: '/plans', method: 'POST', body }),
			invalidatesTags: (_result, error) => error ? [] : [{ type: 'Plan', id: 'LIST' }],
		}),
		updatePlan: build.mutation<BusinessPlan, { id: string; body: PlanPatch }>({
			query: ({ id, body }) => ({ url: `/plans/${id}`, method: 'PATCH', body }),
			invalidatesTags: (_result, error) => error ? [] : [{ type: 'Plan', id: 'LIST' }],
		}),
		deletePlan: build.mutation<void, string>({
			query: (id) => ({ url: `/plans/${id}`, method: 'DELETE' }),
			invalidatesTags: (_result, error) => error ? [] : [{ type: 'Plan', id: 'LIST' }],
		}),
		getPlans: build.query<PlansResponse, PlanState | void>({
			query: (state) => ({ url: '/plans', params: state ? { state } : undefined }),
			providesTags: [{ type: 'Plan', id: 'LIST' }],
		}),
		getPlanRecommendations: build.query<PlanHomeRecommendation[], void>({
			query: () => '/plans/recommendations',
			providesTags: [{ type: 'Plan', id: 'LIST' }],
		}),
	}),
});

export const {
	useGetPlanQuery,
	useGetPlansQuery,
	useGetPlanRecommendationsQuery,
	useCreatePlanMutation,
	useUpdatePlanMutation,
	useDeletePlanMutation,
} = planApi;
