export type UserRole = 'CLIENT' | 'BANK_OFFICER';

/** Пользователь из ответа POST /auth/login */
export interface SessionUser {
	id: string;
	fullName: string;
	position: string;
	role: UserRole;
	companyId: string | null;
}

export interface LoginRequest {
	login: string;
	password: string;
}

export interface LoginResponse {
	user: SessionUser;
	/** Куда перейти после входа: `/` для клиента, `/bank/queue` для сотрудника */
	redirect: string;
}

export interface CompanyRef {
	id: string;
	inn: string;
	name: string;
	isBankClient: boolean;
}

/** Ответ GET /auth/me */
export interface CurrentUser {
	id: string;
	fullName: string;
	position: string;
	role: UserRole;
	company?: CompanyRef | null;
}
