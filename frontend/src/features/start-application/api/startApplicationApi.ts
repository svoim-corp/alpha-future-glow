import { baseApi } from '@shared/api';

export interface CreateApplicationArgs {
	productCode: string;
	amount: number;
	term: number | null;
	goal?: string;
}

export interface CreateApplicationResult {
	id: string;
	status: 'DRAFT';
	version: number;
	redirect: string;
}

export const startApplicationApi = baseApi.injectEndpoints({
	endpoints: (build) => ({
		createApplication: build.mutation<CreateApplicationResult, CreateApplicationArgs>({
			query: (body) => ({ url: '/applications', method: 'POST', body }),
			invalidatesTags: [{ type: 'Application', id: 'LIST' }],
		}),
	}),
});

export const { useCreateApplicationMutation } = startApplicationApi;
