import { baseApi, type Paginated } from '@shared/api';

import {
	type Assignee,
	type BankApplicationDetails,
	type BankQueueFilters,
	type BankQueueItem,
	type BankStats,
	type Completeness,
	type ExternalCheck,
	type HistoryEntry,
	type Score,
} from '../model/bank';
import { type InfoRequest } from '../model/tracking';

const toQuery = (filters: BankQueueFilters): string => {
	const params = new URLSearchParams();

	for (const [key, value] of Object.entries(filters)) {
		if (value !== undefined && value !== '') {
			params.set(key, String(value));
		}
	}

	params.set('sort', '-submittedAt');

	return params.toString();
};

const detailTag = (_result: unknown, _error: unknown, id: string) => [{ type: 'BankApplication' as const, id }];

export const bankApi = baseApi.injectEndpoints({
	endpoints: (build) => ({
		getBankStats: build.query<BankStats, void>({
			query: () => '/bank/applications/stats',
			providesTags: [{ type: 'BankApplication', id: 'LIST' }],
		}),
		getBankApplications: build.query<Paginated<BankQueueItem>, BankQueueFilters>({
			query: (filters) => `/bank/applications?${toQuery(filters)}`,
			providesTags: [{ type: 'BankApplication', id: 'LIST' }],
		}),
		assignApplication: build.mutation<{ assignee: Assignee }, string>({
			query: (id) => ({ url: `/bank/applications/${id}/assign`, method: 'POST' }),
			invalidatesTags: (_result, _error, id) => [
				{ type: 'BankApplication', id: 'LIST' },
				{ type: 'BankApplication', id },
			],
		}),
		getBankApplication: build.query<BankApplicationDetails, string>({
			query: (id) => `/bank/applications/${id}`,
			providesTags: detailTag,
		}),
		getCompleteness: build.query<Completeness, string>({
			query: (id) => `/bank/applications/${id}/completeness`,
			providesTags: detailTag,
		}),
		getChecks: build.query<ExternalCheck[], string>({
			query: (id) => `/bank/applications/${id}/checks`,
			providesTags: detailTag,
		}),
		refreshChecks: build.mutation<ExternalCheck[], { id: string; sources?: string[] }>({
			query: ({ id, sources }) => ({
				url: `/bank/applications/${id}/checks/refresh`,
				method: 'POST',
				body: sources ? { sources } : {},
			}),
			invalidatesTags: (_result, _error, { id }) => [{ type: 'BankApplication', id }],
		}),
		getScore: build.query<Score, string>({
			query: (id) => `/bank/applications/${id}/score`,
			providesTags: detailTag,
		}),
		getBankInfoRequests: build.query<InfoRequest[], string>({
			query: (id) => `/bank/applications/${id}/info-requests`,
			providesTags: detailTag,
		}),
		getBankHistory: build.query<HistoryEntry[], string>({
			query: (id) => `/bank/applications/${id}/history`,
			providesTags: detailTag,
		}),
	}),
});

export const {
	useGetBankStatsQuery,
	useGetBankApplicationsQuery,
	useAssignApplicationMutation,
	useGetBankApplicationQuery,
	useGetCompletenessQuery,
	useGetChecksQuery,
	useRefreshChecksMutation,
	useGetScoreQuery,
	useGetBankHistoryQuery,
	useGetBankInfoRequestsQuery,
} = bankApi;
