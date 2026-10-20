export type ApplicationStatus =
	| 'DRAFT'
	| 'SUBMITTED'
	| 'DATA_CHECK'
	| 'INFO_REQUESTED'
	| 'SCORING'
	| 'ON_DECISION'
	| 'APPROVED'
	| 'COUNTER_OFFER'
	| 'REJECTED'
	| 'SIGNING'
	| 'ISSUED'
	| 'CLOSED';

export interface ProductRef {
	code: string;
	name: string;
}

/** Строка в списках заявок (ApplicationShort, раздел 3 API_SPEC) */
export interface ApplicationShort {
	id: string;
	/** Присваивается при отправке, у черновика — null */
	number: string | null;
	product: ProductRef;
	amount: number;
	term: number;
	status: ApplicationStatus;
	statusLabel: string;
	actionRequired: boolean;
	updatedAt: string;
}

export type ApplicationStage = 1 | 2;
export type CompanyType = 'LLC' | 'IE';
export type AuthorityType = 'HEAD' | 'MCHD';
export type PrelimOutcome = 'LIMIT' | 'STOP';
export type PrelimChoice = 'STATEMENT' | 'REDUCE_AMOUNT';

export interface ApplicationParams {
	amount: number;
	term: number | null;
	goal: string | null;
	purposeText: string | null;
	graceMonths: number | null;
	overdraftLimit: number | null;
}

export interface ApplicationCompany {
	inn: string;
	type: CompanyType;
	name: string;
	ogrn: string;
	regDate: string;
	state: string;
}

export interface ApplicationAuthority {
	type: AuthorityType;
	verified: boolean;
}

export interface ConsentRecord {
	given: boolean;
	at: string;
}

export interface ApplicationPrelim {
	outcome: PrelimOutcome;
	limitNoStatement: number | null;
	limitWithStatement: number | null;
	prelimRate: number | null;
	choice: PrelimChoice | null;
}

export interface ActualAddress {
	sameAsRegistration: boolean;
	address: string | null;
}

export interface TurnoverInfo {
	source: string;
	avgMonthlyIncome: number;
	periodMonths: number;
}

export interface Guarantor {
	fullName: string;
	inn: string;
	bkiConsent: boolean;
}

export interface ApplicationStage2 {
	actualAddress: ActualAddress | null;
	turnover: TurnoverInfo | null;
	guarantor: Guarantor | null;
	truthConfirmed: boolean;
}

export interface ApplicationLocks {
	paramsLocked: boolean;
	stage2Available: boolean;
}

/** Полное состояние заявки: GET /applications/{id} */
export interface Application {
	id: string;
	number: string | null;
	status: ApplicationStatus;
	stage: ApplicationStage;
	version: number;
	product: ProductRef;
	params: ApplicationParams;
	company: ApplicationCompany | null;
	authority: ApplicationAuthority | null;
	consents: { bki?: ConsentRecord };
	prelim: ApplicationPrelim | null;
	stage2: ApplicationStage2;
	locks: ApplicationLocks;
	savedAt: string | null;
}

/** Тело PATCH /applications/{id}: только изменённые поля */
export interface ApplicationPatch {
	params?: Partial<Pick<ApplicationParams, 'amount' | 'term' | 'goal' | 'purposeText' | 'graceMonths' | 'overdraftLimit'>>;
	stage2?: {
		actualAddress?: { sameAsRegistration: boolean; address?: string | null };
		guarantor?: { fullName: string; inn: string };
	};
}

export interface PatchApplicationArgs {
	id: string;
	version: number;
	body: ApplicationPatch;
}

export interface PatchApplicationResult {
	version: number;
	savedAt: string;
	/** true — изменились сумма, срок или продукт: этап 1 нужно пройти заново */
	prelimInvalidated: boolean;
}
