export { dashboardApi, useGetDashboardQuery, useSaveDashboardMutation } from './api/dashboardApi';
export { readLegacyDashboardConfig } from './lib/legacyConfig';
export { createDashboardWidget, createDefaultDashboard } from './model/defaults';
export { DASHBOARD_DEFINITIONS } from './model/definitions';
export type { DashboardDocument, DashboardScope, SaveDashboardArgs } from './model/types';
export { dashboardConfigSchema, dashboardWidgetSchema } from './model/schema';
export type { DashboardConfig, DashboardWidget, DashboardWidgetKind, QuickAction } from './model/schema';
