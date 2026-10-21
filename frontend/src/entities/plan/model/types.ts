export type DateMode = 'DAY' | 'PERIOD' | 'UNSCHEDULED';
export type PlanState = 'ACTIVE' | 'PAUSED' | 'DONE' | 'ARCHIVED';

export interface PlanRecommendation {
	productCode: string;
	productName: string;
	reason: string;
	availableFrom: string;
}

/** Модель BusinessPlan из API_SPEC_ADDITIONAL.md */
export interface BusinessPlan {
	id: string;
	title: string;
	note: string | null;
	goal: string;
	budgetRub: number;
	dateMode: DateMode;
	startDate: string | null;
	endDate: string | null;
	state: PlanState;
	notifications: { inApp: boolean; email: boolean };
	recommendation: PlanRecommendation | null;
	createdAt: string;
}

export interface PlansResponse {
	items: BusinessPlan[];
	summary: { total: number; withPeriod: number; unscheduled: number; budgetRub: number };
}

export interface PlanHomeRecommendation {
	planId: string;
	planTitle: string;
	budgetRub: number;
	productCode: string;
	productName: string;
	reason: string;
}

export interface CreatePlanArgs {
	title: string;
	note: string | null;
	goal: string;
	budgetRub: number;
	dateMode: DateMode;
	startDate: string | null;
	endDate: string | null;
	notifications: { inApp: boolean; email: boolean };
}

export interface PlanPatch extends Partial<Omit<CreatePlanArgs, 'notifications'>> {
	state?: PlanState;
	notifications?: { inApp?: boolean; email?: boolean };
}
