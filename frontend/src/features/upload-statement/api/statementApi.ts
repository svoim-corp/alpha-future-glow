import { baseApi } from '@shared/api';

export interface StatementResult {
	parsed: boolean;
	period: { from: string; to: string };
	operations: number;
	avgMonthlyIncome: number;
	newLimit: number;
	amountFits: boolean;
}

export const statementApi = baseApi.injectEndpoints({
	endpoints: (build) => ({
		uploadStatement: build.mutation<StatementResult, { applicationId: string; file: File }>({
			query: ({ applicationId, file }) => {
				const body = new FormData();
				body.append('file', file);

				return { url: `/applications/${applicationId}/statement`, method: 'POST', body };
			},
			invalidatesTags: (_result, _error, { applicationId }) => [{ type: 'Application', id: applicationId }],
		}),
	}),
});

export const { useUploadStatementMutation } = statementApi;
