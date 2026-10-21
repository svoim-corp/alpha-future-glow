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
export type PrelimChoice = 'REDUCE_AMOUNT' | 'KEEP_AMOUNT';

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
	validUntil?: string | null;
}

export interface ConsentRecord {
	given: boolean;
	at: string;
}

export interface ApplicationPrelim {
	outcome: PrelimOutcome;
	limit: number | null;
	prelimRate: number | null;
	/** false — запрошенная сумма больше лимита, клиент выбирает дальше */
	amountFits: boolean | null;
	choice: PrelimChoice | null;
}

export interface TurnoverInfo {
	source: string;
	avgMonthlyIncome: number;
	periodMonths: number;
}

export interface Guarantor {
	personId: string;
	fullName: string;
	bkiConsent: boolean;
}

export interface ApplicationStage2 {
	/** Данные банка, только для чтения */
	actualAddress: string | null;
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
		guarantor?: { personId: string };
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

export interface PersonRef {
	personId: string;
	fullName: string;
}

export type PersonRole = 'HEAD' | 'FOUNDER';

export interface GuarantorCandidate extends PersonRef {
	roles: PersonRole[];
}

/** Данные банка, сверенные с реестрами, только для чтения: GET /applications/{id}/company-profile */
export interface CompanyProfile {
	source: string[];
	regAddress: string;
	actualAddress: string | null;
	okved: { code: string; name: string };
	head: PersonRef;
	founders: (PersonRef & { share: number })[];
	/** Руководитель и учредители-физлица; для ИП — пустой массив */
	guarantorCandidates: GuarantorCandidate[];
	taxRegime: string;
	turnover: TurnoverInfo | null;
	girbo: { accounted: boolean } | null;
}
