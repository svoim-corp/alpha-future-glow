export interface Conversation {
	id: string;
	applicationId: string | null;
	createdAt: string;
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
	kind: 'draft' | 'ticket';
	draft?: ChatDraft;
	subject?: string;
	text?: string;
}

export interface ChatBlock {
	kind: 'product' | 'calculation' | 'application' | 'payment' | 'action' | 'result';
	title: string;
	product?: ChatProduct;
	calculation?: ChatCalculation;
	application?: ChatApplication;
	payment?: ChatPayment;
	actionId?: string;
	action?: ChatAction;
	resultId?: string;
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
