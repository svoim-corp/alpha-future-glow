import { baseApi } from '@shared/api';

import { type Contract } from '../model/types';

export const contractApi = baseApi.injectEndpoints({
	endpoints: (build) => ({
		getContract: build.query<Contract, string>({
			query: (applicationId) => `/applications/${applicationId}/contract`,
			providesTags: (_result, _error, id) => [{ type: 'Application', id }],
		}),
	}),
});

export const { useGetContractQuery } = contractApi;
