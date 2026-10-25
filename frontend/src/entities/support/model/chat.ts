export interface Conversation {
	id: string;
	applicationId: string | null;
	createdAt: string;
	title?: string;
}

export interface ChatProduct {
	code: string;
	name: string;
	goals: string[];
	amountMin: number;
	amountMax: number;
	termMin: number;
	termMax: number;
	rateMin: number;
	rateMax: number;
}

export interface ChatApplication {
	id: string;
	number: string | null;
	productCode: string;
	amount: number;
	term: number;
	status: string;
	nextStep: string;
	canDelete?: boolean;
}

export interface ChatCalculation {
	amount: number;
	term: number;
	min: number;
	max: number;
	rateMin: number;
	rateMax: number;
}

export interface ChatPayment {
	loanId: string;
	productName: string;
	date: string;
	amount: number;
}

export interface ChatDraft {
	productCode: string;
	amount: number;
	term: number;
	goal: string;
}

export interface ChatAction {
	productName?: string;
	goalLabel?: string;
	kind: 'draft' | 'ticket' | 'plan';
	draft?: ChatDraft;
	subject?: string;
	text?: string;
	planPreview?: ChatPlan;
	planChange?: Partial<ChatPlan>;
}

export interface ChatPlan {
	id: string;
	title: string;
	goal: string;
	note?: string;
	budgetRub: number;
	dateMode: 'DAY' | 'PERIOD' | 'UNSCHEDULED';
	startDate?: string;
	endDate?: string;
	state: 'ACTIVE' | 'PAUSED' | 'DONE' | 'ARCHIVED';
	updatedAt: string;
}

export interface ChatBlock {
	kind: 'product' | 'calculation' | 'application' | 'payment' | 'action' | 'result'
		| 'navigation' | 'plan' | 'loan' | 'company' | 'ticket';
	title: string;
	product?: ChatProduct;
	calculation?: ChatCalculation;
	application?: ChatApplication;
	payment?: ChatPayment;
	actionId?: string;
	action?: ChatAction;
	resultId?: string;
	navigation?: { section: string; label: string; href: string };
	plan?: ChatPlan;
	loan?: {
		id: string; productName: string; amount: number; balance: number; rate: number; term: number; status: string;
	};
	company?: {
		name: string; inn: string; regAddress: string; actualAddress: string; monthlyIncome: number;
		viewerName?: string; viewerPosition?: string;
	};
	ticket?: { id: string; subject: string; text: string; status: string; reply?: string };
}

export interface ChatMessage {
	id: string;
	role: 'user' | 'assistant';
	status: 'PENDING' | 'COMPLETED' | 'FAILED';
	text: string;
	sources: { id: string; title: string; href: string; excerpt?: string }[];
	blocks: ChatBlock[];
	createdAt: string;
}
