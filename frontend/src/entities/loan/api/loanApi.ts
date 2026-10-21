import { baseApi } from '@shared/api';

import { type Loan, type LoanDetails, type LoansSummary, type ScheduleItem } from '../model/types';

export const loanApi = baseApi.injectEndpoints({
	endpoints: (build) => ({
		getLoansSummary: build.query<LoansSummary, void>({
			query: () => '/loans/summary',
			providesTags: ['Loan'],
		}),
		getLoans: build.query<Loan[], void>({
			query: () => '/loans',
			providesTags: ['Loan'],
		}),
		getLoan: build.query<LoanDetails, string>({
			query: (id) => `/loans/${id}`,
			providesTags: (_result, _error, id) => [{ type: 'Loan', id }],
		}),
		getSchedule: build.query<ScheduleItem[], { from: string; to: string }>({
			query: ({ from, to }) => `/loans/schedule?from=${from}&to=${to}`,
			providesTags: ['Loan'],
		}),
		getUpcomingPayments: build.query<ScheduleItem[], number>({
			query: (limit) => `/loans/schedule/upcoming?limit=${limit}`,
			providesTags: ['Loan'],
		}),
	}),
});

export const {
	useGetLoansSummaryQuery,
	useGetLoansQuery,
	useGetLoanQuery,
	useGetScheduleQuery,
	useGetUpcomingPaymentsQuery,
} = loanApi;
