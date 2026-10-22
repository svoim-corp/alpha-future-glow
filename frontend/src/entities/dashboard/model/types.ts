import { type DashboardConfig } from './schema';

export interface DashboardScope { userId: string; companyId: string }
export interface DashboardDocument { revision: string; updatedAt: string; config: DashboardConfig }
export interface SaveDashboardArgs {
	scope: DashboardScope;
	config: DashboardConfig;
	expectedRevision: string | null;
	idempotencyKey: string;
}
