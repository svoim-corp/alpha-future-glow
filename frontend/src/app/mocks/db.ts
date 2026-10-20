import {
	type Application,
	type ApplicationShort,
	type ApplicationStatus,
	type InfoRequest,
	type Offer,
} from '@entities/application';
import { type Contract } from '@entities/contract';
import { type DocumentsResponse } from '@entities/document';
import { type Loan, type LoanDetails } from '@entities/loan';
import { type NotificationsResponse } from '@entities/notification';
import { type BusinessPlan } from '@entities/plan';
import { type CurrentUser } from '@entities/session';
import { type Ticket } from '@entities/support';

import { draftApplication, fullApplication, infoRequestedApplication } from './fixtures/applications';
import { contractFixture } from './fixtures/contract';
import { documentsFixture } from './fixtures/documents';
import { loanDetailsFixture, loanFixture } from './fixtures/loans';
import { notificationsFixture } from './fixtures/notifications';
import { plansFixture } from './fixtures/plans';
import { clientUser, officerUser } from './fixtures/session';
import { ticketsFixture } from './fixtures/support';
import { openInfoRequest } from './fixtures/tracking';

const SESSION_KEY = 'mock-session';

export const STATUS_LABELS: Record<ApplicationStatus, string> = {
	DRAFT: 'Черновик',
	SUBMITTED: 'Подана',
	DATA_CHECK: 'Проверка данных',
	INFO_REQUESTED: 'Нужна доп. информация',
	SCORING: 'Обогащение и скоринг',
	ON_DECISION: 'На решении',
	APPROVED: 'Одобрена',
	COUNTER_OFFER: 'Встречное предложение',
	REJECTED: 'Отказ',
	SIGNING: 'Подписание',
	ISSUED: 'Кредит выдан',
	CLOSED: 'Закрыта',
};

export interface MockApplication {
	short: ApplicationShort;
	full: Application;
	infoRequests: InfoRequest[];
	offer: Offer | null;
	/** Когда заявка была подана (для истории и очереди банка) */
	submittedAt: string | null;
	assignee: { id: string; name: string } | null;
	loanId: string | null;
}

export interface MockDb {
	applications: MockApplication[];
	documents: DocumentsResponse['items'];
	plans: BusinessPlan[];
	tickets: Ticket[];
	notifications: NotificationsResponse['items'];
	loans: { loan: Loan; details: LoanDetails }[];
	contract: Contract;
	seq: number;
}

const clone = <T>(value: T): T => structuredClone(value);

const createDb = (): MockDb => {
	const infoRequested: MockApplication = {
		short: clone(infoRequestedApplication),
		full: { ...clone(fullApplication), id: 'a-1024', number: infoRequestedApplication.number, status: 'INFO_REQUESTED' },
		infoRequests: [clone(openInfoRequest)],
		offer: null,
		submittedAt: '2026-10-20T10:25:00Z',
		assignee: null,
		loanId: null,
	};
	const draft: MockApplication = {
		short: clone(draftApplication),
		full: clone(fullApplication),
		infoRequests: [],
		offer: null,
		submittedAt: null,
		assignee: null,
		loanId: null,
	};

	return {
		applications: [infoRequested, draft],
		documents: clone(documentsFixture.items),
		plans: clone(plansFixture.items),
		tickets: clone(ticketsFixture),
		notifications: clone(notificationsFixture.items),
		loans: [{ loan: clone(loanFixture), details: clone(loanDetailsFixture) }],
		contract: clone(contractFixture),
		seq: 2000,
	};
};

export const db: MockDb = createDb();

export const nextId = (prefix: string) => {
	db.seq += 1;

	return `${prefix}-${db.seq}`;
};

export const findApplication = (id: string) => db.applications.find((item) => item.short.id === id);

/** Сессия переживает перезагрузку страницы, но не закрытие вкладки */
export const session = {
	get(): CurrentUser | null {
		try {
			const raw = sessionStorage.getItem(SESSION_KEY);

			return raw === 'CLIENT' ? clientUser : raw === 'BANK_OFFICER' ? officerUser : null;
		} catch {
			return null;
		}
	},
	set(role: 'CLIENT' | 'BANK_OFFICER' | null) {
		try {
			if (role) {
				sessionStorage.setItem(SESSION_KEY, role);
			} else {
				sessionStorage.removeItem(SESSION_KEY);
			}
		} catch {
			// без sessionStorage вход не переживёт перезагрузку
		}
	},
};

export const setStatus = (item: MockApplication, status: ApplicationStatus) => {
	item.short.status = status;
	item.short.statusLabel = STATUS_LABELS[status];
	item.short.actionRequired = status === 'INFO_REQUESTED' || status === 'APPROVED' || status === 'COUNTER_OFFER'
		|| status === 'SIGNING';
	item.short.updatedAt = new Date().toISOString();
	item.full.status = status;
};
