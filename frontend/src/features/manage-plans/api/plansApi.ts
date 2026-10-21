import { baseApi } from '@shared/api';

export interface CreatePlanArgs {
	title: string;
	note: string | null;
	goal: string | null;
	budgetRub: number;
	dateMode: 'DAY' | 'PERIOD' | 'UNSCHEDULED';
	startDate: string | null;
	endDate: string | null;
	notifications: { inApp: boolean; email: boolean };
}

export const plansApi = baseApi.injectEndpoints({
	endpoints: (build) => ({
		createPlan: build.mutation<unknown, CreatePlanArgs>({
			query: (body) => ({ url: '/plans', method: 'POST', body }),
			invalidatesTags: [{ type: 'Plan', id: 'LIST' }],
		}),
		deletePlan: build.mutation<void, string>({
			query: (id) => ({ url: `/plans/${id}`, method: 'DELETE' }),
			invalidatesTags: [{ type: 'Plan', id: 'LIST' }],
		}),
	}),
});

export const { useCreatePlanMutation, useDeletePlanMutation } = plansApi;
