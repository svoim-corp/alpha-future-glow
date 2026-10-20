import { type CurrentUser, type LoginResponse } from '@entities/session';

export const clientUser: CurrentUser = {
	id: 'u-1',
	fullName: 'Смирнов Олег Викторович',
	position: 'Генеральный директор',
	role: 'CLIENT',
	company: { id: 'c-1', inn: '7712345678', name: 'ООО «Северная пекарня»', isBankClient: true },
};

export const officerUser: CurrentUser = {
	id: 'u-9',
	fullName: 'Иванова Елена Петровна',
	position: 'Кредитный инспектор',
	role: 'BANK_OFFICER',
	company: null,
};

export const clientLogin: LoginResponse = {
	user: { id: 'u-1', fullName: 'Смирнов Олег Викторович', position: 'Генеральный директор', role: 'CLIENT', companyId: 'c-1' },
	redirect: '/',
};

export const officerLogin: LoginResponse = {
	user: { id: 'u-9', fullName: 'Иванова Елена Петровна', position: 'Кредитный инспектор', role: 'BANK_OFFICER', companyId: null },
	redirect: '/bank/queue',
};
