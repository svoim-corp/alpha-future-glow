import { type ApplicationStatus, type ProductRef } from './types';

export type RiskBand = 'LOW' | 'MEDIUM' | 'HIGH';
export type DecisionMode = 'AUTO' | 'MANUAL';

/** GET /bank/applications/stats */
export interface BankStats {
	SUBMITTED: number;
	DATA_CHECK: number;
	INFO_REQUESTED: number;
	SCORING: number;
	ON_DECISION: number;
	decidedToday: number;
}

export interface Assignee {
	id: string;
	fullName: string;
}

export interface BankQueueItem {
	id: string;
	number: string;
	company: { inn: string; name: string };
	product: ProductRef;
	amount: number;
	term: number;
	status: ApplicationStatus;
	score: number | null;
	riskBand: RiskBand | null;
	decisionMode: DecisionMode;
	assignee: Assignee | null;
	submittedAt: string;
}

export interface BankQueueFilters {
	status?: ApplicationStatus;
	product?: string;
	risk?: RiskBand;
	decisionMode?: DecisionMode;
	page: number;
	size: number;
}

export interface BankApplicationDetails {
	id: string;
	number: string;
	status: ApplicationStatus;
	applicant: { fullName: string; authority: 'HEAD' | 'MCHD'; mchdValidUntil?: string };
	company: {
		inn: string;
		name: string;
		type: 'LLC' | 'IE';
		regDate: string;
		okved: string;
		taxRegime: string;
		actualAddress: string | null;
	};
	request: {
		productCode: string;
		amount: number;
		term: number;
		goal: string | null;
		purposeText: string | null;
	};
	turnover: { source: string; avgMonthlyIncome: number } | null;
	girbo: { year: number; revenue: number; netProfit: number } | null;
	guarantor: { fullName: string; inn: string; bkiConsent: boolean } | null;
	prelim: { limitNoStatement: number; limitWithStatement: number; prelimRate: number } | null;
}

export type IssueSeverity = 'INFO' | 'WARN' | 'BLOCKER';

export interface Completeness {
	complete: boolean;
	issues: { severity: IssueSeverity; code: string; text: string }[];
}

export type CheckStatus = 'OK' | 'WARN' | 'FAIL' | 'UNAVAILABLE';

export interface ExternalCheck {
	source: string;
	status: CheckStatus;
	requestedAt: string;
	summary: string;
	details?: Record<string, unknown>;
}

export interface ScoreFactor {
	code: string;
	label: string;
	impact: number;
}

export interface Score {
	score: number;
	riskBand: RiskBand;
	autoDecisionEligible: boolean;
	autoDecisionReason: string | null;
	factors: ScoreFactor[];
	suggested: { decision: 'APPROVE' | 'COUNTER_OFFER' | 'REJECT'; amount?: number; rate?: number } | null;
	modelVersion: string;
}

export interface HistoryEntry {
	at: string;
	actor: { type: 'CLIENT' | 'SYSTEM' | 'OFFICER'; name?: string };
	action: string;
	from: ApplicationStatus;
	to: ApplicationStatus;
	comment?: string;
}
