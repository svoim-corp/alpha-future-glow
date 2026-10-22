import { baseApi } from '@shared/api';
import { z } from 'zod';

import { dashboardConfigSchema } from '../model/schema';
import { type DashboardDocument, type DashboardScope, type SaveDashboardArgs } from '../model/types';

const documentSchema = z.object({
	revision: z.string().min(1).max(128).regex(/^[^\s"\\]+$/u),
	updatedAt: z.iso.datetime(),
	config: dashboardConfigSchema,
}).strict();
const unsupportedDocument = () => ({
	status: 'CUSTOM_ERROR' as const,
	error: 'Unsupported dashboard document',
	data: { code: 'DASHBOARD_SCHEMA_UNSUPPORTED', message: 'Сервер вернул неподдерживаемый формат настройки. Обновите страницу, чтобы загрузить актуальную версию.' },
});
const scopeKey = ({ userId, companyId }: DashboardScope) => JSON.stringify([userId, companyId]);

export const dashboardApi = baseApi.injectEndpoints({
	endpoints: (build) => ({
		getDashboard: build.query<DashboardDocument | null, DashboardScope>({
			async queryFn(_scope, _api, _options, baseQuery) {
				const result = await baseQuery({ url: '/me/dashboard', cache: 'no-store' });
				if (result.error) return { error: result.error };
				if (result.meta?.response?.status === 204) return { data: null };
				const parsed = documentSchema.safeParse(result.data);
				return parsed.success ? { data: parsed.data } : { error: unsupportedDocument() };
			},
			providesTags: (_result, _error, scope) => [{ type: 'Dashboard', id: scopeKey(scope) }],
			keepUnusedDataFor: 0,
		}),
		saveDashboard: build.mutation<DashboardDocument, SaveDashboardArgs>({
			async queryFn({ config, expectedRevision, idempotencyKey }, _api, _options, baseQuery) {
				const result = await baseQuery({
					url: '/me/dashboard', method: 'PUT', body: config,
					headers: {
						...(expectedRevision === null ? { 'If-None-Match': '*' } : { 'If-Match': `"${expectedRevision}"` }),
						'Idempotency-Key': idempotencyKey,
					},
				});
				if (result.error) return { error: result.error };
				const parsed = documentSchema.safeParse(result.data);
				return parsed.success ? { data: parsed.data } : { error: unsupportedDocument() };
			},
			invalidatesTags: (_result, error, { scope }) => error ? [] : [{ type: 'Dashboard', id: scopeKey(scope) }],
		}),
	}),
});

export const { useGetDashboardQuery, useSaveDashboardMutation } = dashboardApi;
