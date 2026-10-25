import { baseApi, type Paginated } from '@shared/api';

import { type Loan, type LoanDebt, type LoanDetails, type LoanRepayment, type LoansSummary, type ScheduleItem } from '../model/types';

export const loanApi = baseApi.injectEndpoints({
	endpoints: (build) => ({
		getLoansSummary: build.query<LoansSummary, void>({
			query: () => '/loans/summary',
			providesTags: ['Loan'],
		}),
		getLoans: build.query<Loan[], void>({
			/** Полный список для выбора договора и переходов из виджетов. */
			async queryFn(_arg, _api, _options, baseQuery) {
				const items = new Map<string, Loan>();
				let expectedTotal: number | undefined;
				for (let page = 1; page <= 1000; page += 1) {
					const result = await baseQuery({ url: '/loans/page', params: { page, size: 100 } });
					if (result.error) return { error: result.error };
					const data = result.data as Paginated<Loan> | null;
					if (!data || !Array.isArray(data.items) || data.page !== page || !Number.isInteger(data.total) || data.total < 0
						|| !Number.isInteger(data.size) || data.size < 1 || data.size > 100 || data.items.length > data.size
						|| (expectedTotal !== undefined && expectedTotal !== data.total)) {
						return { error: { status: 'CUSTOM_ERROR', error: 'Список кредитов изменился или получена некорректная страница. Повторите загрузку.' } };
					}
					expectedTotal = data.total;
					for (const item of data.items) items.set(item.id, item);
					if (page * data.size >= data.total) {
						if (items.size === data.total) return { data: [...items.values()] };
						break;
					}
					if (data.items.length === 0) break;
				}
				return { error: { status: 'CUSTOM_ERROR', error: 'Не удалось загрузить полный список кредитов. Повторите загрузку.' } };
			},
			providesTags: ['Loan'],
		}),
		getLoan: build.query<LoanDetails, string>({
			query: (id) => `/loans/${id}`,
			providesTags: (_result, _error, id) => [{ type: 'Loan', id }],
		}),
		getLoanDebt: build.query<LoanDebt, string>({
			query: (id) => `/loans/${id}/debt`,
			providesTags: (_result, _error, id) => [{ type: 'Loan', id }],
		}),
		getLoanRepayments: build.query<Paginated<LoanRepayment>, { loanId: string; page: number }>({
			query: ({ loanId, page }) => ({ url: `/loans/${loanId}/repayments`, params: { page, size: 20 } }),
			providesTags: (_result, _error, { loanId }) => [{ type: 'Loan', id: loanId }],
		}),
		// GET с жизненным циклом mutation: каждое нажатие скачивает свежий файл, Blob не попадает в Redux.
		downloadLoanSchedule: build.mutation<null, string>({
			async queryFn(id, _api, _options, baseQuery) {
				const result = await baseQuery({
					url: `/loans/${id}/schedule/document`,
					params: { format: 'pdf' },
					responseHandler: async (response): Promise<unknown> => {
						if (response.ok) return response.blob();
						const errorBody: unknown = await response.json();
						return errorBody;
					},
				});
				if (result.error) return { error: result.error };
				if (!(result.data instanceof Blob) || result.data.type.split(';')[0] !== 'application/pdf') {
					return { error: { status: 'CUSTOM_ERROR', error: 'Сервер не вернул PDF графика. Повторите загрузку.' } };
				}
				const url = URL.createObjectURL(result.data);
				const link = document.createElement('a');
				link.href = url;
				link.download = `schedule-${id}.pdf`;
				document.body.append(link);
				link.click();
				link.remove();
				setTimeout(() => URL.revokeObjectURL(url), 1000);
				return { data: null };
			},
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
	useGetLoanDebtQuery,
	useGetLoanRepaymentsQuery,
	useDownloadLoanScheduleMutation,
	useGetScheduleQuery,
	useGetUpcomingPaymentsQuery,
} = loanApi;
