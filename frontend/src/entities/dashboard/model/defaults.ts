import { DASHBOARD_DEFINITIONS } from './definitions';
import { type DashboardConfig, type DashboardWidget, type DashboardWidgetKind } from './schema';

export const createDashboardWidget = (type: DashboardWidgetKind, id: string): DashboardWidget => {
	const [w, h] = DASHBOARD_DEFINITIONS[type].sizes[0];
	const base = { id, x: 0, y: 0, w, h };
	switch (type) {
		case 'applications': return { ...base, type, settings: { filter: 'all' } };
		case 'loan': return { ...base, type, settings: { loanId: null } };
		case 'payment': return { ...base, type, settings: { loanId: null } };
		case 'debt': return { ...base, type, settings: {} };
		case 'quick': return { ...base, type, settings: { actions: ['apply', 'plan', 'support'] } };
		case 'calendar': return { ...base, type, settings: { source: 'all' } };
		case 'plans': return { ...base, type, settings: { filter: 'upcoming' } };
		case 'application': return { ...base, type, settings: { applicationId: null } };
		case 'loans': return { ...base, type, settings: { status: 'ACTIVE' } };
		case 'payments': return { ...base, type, settings: { days: 30, loanId: null } };
		case 'offers': return { ...base, type, settings: { filter: 'current' } };
		case 'drafts': return { ...base, type, settings: {} };
		case 'planFunding': return { ...base, type, settings: { planId: null } };
		case 'notifications': return { ...base, type, settings: { filter: 'unread' } };
		case 'tickets': return { ...base, type, settings: { filter: 'active' } };
		case 'funding': return { ...base, type, settings: {} };
	}
};

export const createDefaultDashboard = (): DashboardConfig => ({
	schemaVersion: 1,
	widgets: [
		{ ...createDashboardWidget('applications', 'default-applications'), x: 0, y: 0 },
		{ ...createDashboardWidget('loan', 'default-loan'), x: 8, y: 0 },
		{ ...createDashboardWidget('payment', 'default-payment'), x: 0, y: 3 },
		{ ...createDashboardWidget('debt', 'default-debt'), x: 4, y: 3 },
		{ ...createDashboardWidget('quick', 'default-quick'), x: 8, y: 3 },
		{ ...createDashboardWidget('calendar', 'default-calendar'), x: 0, y: 5 },
		{ ...createDashboardWidget('plans', 'default-plans'), x: 6, y: 5 },
		{ ...createDashboardWidget('application', 'default-application'), x: 0, y: 9 },
	],
});
