import {
	type Application,
	type ApplicationStatus,
	type Offer,
	type Timeline,
} from '@entities/application';
import { env } from '@shared/config';
import { http, HttpResponse } from 'msw';

import { db, findApplication, type MockApplication, nextId, session, setStatus, STATUS_LABELS } from './db';
import {
	bankDetailsFixture,
	bankStatsFixture,
	checksFixture,
	completenessFixture,
	historyFixture,
	scoreFixture,
} from './fixtures/bank';
import { companyProfileFixture } from './fixtures/companyProfile';
import { scheduleFixture } from './fixtures/loans';
import {
	goalsFixture,
	investDetails,
	matchFixture,
	overdraftDetails,
	productsFixture,
	workingDetails,
} from './fixtures/products';
import { clientLogin, officerLogin } from './fixtures/session';
import { faqFixture } from './fixtures/support';
import { plannerHandlers } from './planner';

const url = (path: string) => `${env.apiBaseUrl}${path}`;

const fail = (status: number, code: string, message: string) => HttpResponse.json(
	{ code, message, fields: [], traceId: 'mock-trace' },
	{ status },
);
const notFound = () => fail(404, 'NOT_FOUND', 'Не найдено');
const nowIso = () => new Date().toISOString();

const STEPS: ApplicationStatus[] = ['SUBMITTED', 'DATA_CHECK', 'INFO_REQUESTED', 'SCORING', 'ON_DECISION'];
const WHATS: Partial<Record<ApplicationStatus, [string, string, 'CLIENT' | 'BANK']>> = {
	SUBMITTED: ['Заявка принята банком', 'Банк начнёт проверку данных', 'BANK'],
	DATA_CHECK: ['Банк проверяет данные компании', 'Дождитесь результата проверки', 'BANK'],
	INFO_REQUESTED: ['Банк просит уточнить данные', 'Ответьте на запрос банка', 'CLIENT'],
	SCORING: ['Идёт скоринг', 'Дождитесь решения', 'BANK'],
	ON_DECISION: ['Заявка передана на решение', 'Дождитесь решения', 'BANK'],
	APPROVED: ['Заявка одобрена', 'Ознакомьтесь с предложением банка', 'CLIENT'],
	COUNTER_OFFER: ['Банк предложил другие условия', 'Примите или отклоните предложение', 'CLIENT'],
	REJECTED: ['Банк отказал в кредите', 'Посмотрите альтернативы в каталоге', 'CLIENT'],
	SIGNING: ['Договор готов к подписанию', 'Подпишите договор', 'CLIENT'],
	ISSUED: ['Кредит выдан', 'Следите за графиком платежей', 'BANK'],
};

const timelineOf = (item: MockApplication): Timeline => {
	const status = item.short.status;
	const index = STEPS.indexOf(status);
	const [whatsHappening, nextStep, actor] = WHATS[status] ?? ['Заявка в работе', 'Дождитесь решения', 'BANK'];

	return {
		current: { status, label: STATUS_LABELS[status], whatsHappening, nextStep, actor },
		steps: STEPS.map((step, position) => ({
			status: step,
			label: STATUS_LABELS[step],
			at: index === -1 || position <= index ? item.submittedAt : null,
			state: index === -1 || position < index ? 'DONE' : position === index ? 'CURRENT' : 'PENDING',
		})),
	};
};

const annuity = (amount: number, term: number, rate: number) => {
	const monthly = rate / 12 / 100;

	return Math.round(monthly === 0 ? amount / term : (amount * monthly) / (1 - (1 + monthly) ** -term));
};

const merge = <T extends object>(target: T, patch: Partial<T>) => {
	for (const [key, value] of Object.entries(patch)) {
		const current = (target as Record<string, unknown>)[key];

		if (value && typeof value === 'object' && !Array.isArray(value) && current && typeof current === 'object') {
			merge(current, value);
		} else {
			(target as Record<string, unknown>)[key] = value;
		}
	}
};

const productDetails = (code: string) => [workingDetails, investDetails, overdraftDetails]
	.find((item) => item.code === code);

export const handlers = [
	// ── сессия ─────────────────────────────────────────────
	http.get(url('/auth/me'), () => {
		const user = session.get();

		return user ? HttpResponse.json(user) : fail(401, 'UNAUTHORIZED', 'Требуется вход');
	}),
	http.post(url('/auth/login'), async ({ request }) => {
		const body = await request.json() as { login: string; password: string };

		if (body.password !== 'demo123') {
			return fail(401, 'INVALID_CREDENTIALS', 'Неверный логин или пароль');
		}

		const officer = /ivanova|officer|bank/i.test(body.login);
		session.set(officer ? 'BANK_OFFICER' : 'CLIENT');

		return HttpResponse.json(officer ? officerLogin : clientLogin);
	}),
	http.post(url('/auth/logout'), () => {
		session.set(null);

		return new HttpResponse(null, { status: 204 });
	}),

	// ── продукты ───────────────────────────────────────────
	http.get(url('/products/goals'), () => HttpResponse.json(goalsFixture)),
	http.get(url('/products'), () => HttpResponse.json(productsFixture)),
	http.post(url('/products/match'), () => HttpResponse.json(matchFixture)),
	http.get(url('/products/:code'), ({ params }) => {
		const details = productDetails(String(params.code));

		return details ? HttpResponse.json(details) : notFound();
	}),
	http.post(url('/calculator/payment'), async ({ request }) => {
		const body = await request.json() as {
			productCode: string;
			amount: number;
			term: number;
			graceMonths?: number;
		};
		const product = productsFixture.find((item) => item.code === body.productCode);
		const rate = product?.rate ?? { min: 20, max: 28 };

		return HttpResponse.json({
			rate,
			payment: {
				min: annuity(body.amount, body.term, rate.min),
				max: annuity(body.amount, body.term, rate.max),
			},
			gracePayment: body.graceMonths
				? {
					months: body.graceMonths,
					min: Math.round((body.amount * rate.min) / 1200),
					max: Math.round((body.amount * rate.max) / 1200),
				}
				: null,
			withinProductLimits: !product
				|| (body.amount >= product.amount.min && body.amount <= product.amount.max),
		});
	}),

	// ── заявки клиента ─────────────────────────────────────
	http.get(url('/applications'), () => {
		const items = db.applications
			.map((item) => item.short)
			.sort((a, b) => b.updatedAt.localeCompare(a.updatedAt));

		return HttpResponse.json({ items, page: 1, size: 20, total: items.length });
	}),
	http.post(url('/applications'), async ({ request }) => {
		const body = await request.json() as {
			productCode: string;
			amount: number;
			term: number | null;
			goal?: string;
		};
		const product = productsFixture.find((item) => item.code === body.productCode);
		const id = nextId('a');
		const full: Application = {
			id,
			number: null,
			status: 'DRAFT',
			stage: 1,
			version: 1,
			product: { code: body.productCode, name: product?.name ?? body.productCode },
			params: {
				amount: body.amount,
				term: body.term ?? 12,
				goal: body.goal ?? null,
				purposeText: null,
				graceMonths: null,
				overdraftLimit: null,
			},
			company: null,
			authority: null,
			consents: {},
			prelim: null,
			stage2: {
				actualAddress: null,
				turnover: null,
				guarantor: null,
				truthConfirmed: false,
			},
			locks: { paramsLocked: false, stage2Available: false },
			savedAt: nowIso(),
		} as unknown as Application;

		db.applications.unshift({
			short: {
				id,
				number: null,
				product: full.product,
				amount: body.amount,
				term: body.term ?? 12,
				status: 'DRAFT',
				statusLabel: STATUS_LABELS.DRAFT,
				actionRequired: false,
				updatedAt: nowIso(),
			},
			full,
			infoRequests: [],
			offer: null,
			submittedAt: null,
			assignee: null,
			loanId: null,
		});

		return HttpResponse.json(
			{ id, status: 'DRAFT', version: 1, redirect: `/applications/${id}/edit` },
			{ status: 201 },
		);
	}),
	http.get(url('/applications/:id'), ({ params }) => {
		const item = findApplication(String(params.id));

		return item ? HttpResponse.json(item.full) : notFound();
	}),
	http.patch(url('/applications/:id'), async ({ params, request }) => {
		const item = findApplication(String(params.id));

		if (!item) {
			return notFound();
		}

		const body = await request.json() as Partial<Application>;
		const before = item.full.params;
		merge(item.full, body);
		const personId = body.stage2?.guarantor?.personId;

		if (personId) {
			const candidate = companyProfileFixture.guarantorCandidates.find((entry) => entry.personId === personId);

			if (!candidate) {
				return HttpResponse.json(
					{ code: 'VALIDATION_ERROR', message: 'Поручитель не из списка кандидатов' },
					{ status: 400 },
				);
			}

			item.full.stage2.guarantor = { personId, fullName: candidate.fullName, bkiConsent: false };
		}

		const changed = before.amount !== item.full.params.amount || before.term !== item.full.params.term;
		const invalidated = Boolean(item.full.prelim) && changed;

		if (invalidated) {
			item.full.prelim = null;
			item.full.stage = 1;
			item.full.locks.stage2Available = false;
		}

		item.full.version += 1;
		item.full.savedAt = nowIso();
		item.short.amount = item.full.params.amount;
		item.short.term = item.full.params.term ?? item.short.term;
		item.short.updatedAt = item.full.savedAt;

		return HttpResponse.json({
			version: item.full.version,
			savedAt: item.full.savedAt,
			prelimInvalidated: invalidated,
		});
	}),
	http.delete(url('/applications/:id'), ({ params }) => {
		db.applications = db.applications.filter((item) => item.short.id !== params.id);

		return new HttpResponse(null, { status: 204 });
	}),
	http.get(url('/applications/:id/company-profile'), () => HttpResponse.json(companyProfileFixture)),
	http.post(url('/applications/:id/company'), async ({ params, request }) => {
		const item = findApplication(String(params.id));
		const body = await request.json() as { inn: string };

		if (!item) {
			return notFound();
		}

		const company = {
			inn: body.inn,
			type: body.inn.length === 12 ? 'IE' as const : 'LLC' as const,
			name: body.inn === '7712345678' ? 'ООО «Северная пекарня»' : `Компания ${body.inn}`,
			ogrn: '1157746000000',
			regDate: '2015-03-12',
			state: 'ACTIVE',
			head: 'Смирнов Олег Викторович',
		};
		item.full.company = { ...company };

		return HttpResponse.json({ company, applicantIsHead: true, authorityRequired: 'NONE' });
	}),
	http.post(url('/applications/:id/authority'), () => HttpResponse.json({
		verified: true,
		principal: 'Смирнов Олег Викторович',
		representative: 'Петрова Анна Сергеевна',
		validUntil: '2027-03-01',
		powers: ['Подписание заявок на кредит'],
	})),
	http.post(url('/applications/:id/consents/bki'), ({ params }) => {
		const item = findApplication(String(params.id));
		const at = nowIso();

		if (item) {
			item.full.consents = { bki: { given: true, at } };
		}

		return HttpResponse.json({ given: true, at });
	}),
	http.post(url('/applications/:id/guarantor/consent-bki'), ({ params }) => {
		const item = findApplication(String(params.id));

		if (item?.full.stage2.guarantor) {
			item.full.stage2.guarantor.bkiConsent = true;
		}

		return HttpResponse.json({ given: true, at: nowIso() }, { status: 201 });
	}),
	http.post(url('/applications/:id/prelim'), ({ params }) => {
		const item = findApplication(String(params.id));

		if (!item) {
			return notFound();
		}

		const { amount } = item.full.params;
		const term = item.full.params.term ?? item.short.term;

		const limit = 3000000;

		if (amount > limit * 2) {
			return HttpResponse.json({
				outcome: 'STOP',
				stopReason: 'CRITICAL_RISK',
				stopText: 'Запрошенная сумма намного выше доступного лимита',
				whatToDo: ['Уменьшите сумму и попробуйте снова', 'Подберите другой продукт в каталоге'],
			});
		}

		const amountFits = amount <= limit;
		const result = {
			outcome: 'LIMIT' as const,
			requestedAmount: amount,
			limit,
			prelimRate: 22,
			prelimPayment: annuity(Math.min(amount, limit), term, 22),
			amountFits,
			checks: [
				{ source: 'EGRUL', status: 'OK', label: 'Компания действует' },
				{ source: 'BANK_ACCOUNTS', status: 'OK', label: 'Обороты по счетам учтены' },
				{ source: 'BKI', status: 'OK', label: 'Кредитная история без просрочек' },
			],
		};
		item.full.prelim = {
			outcome: 'LIMIT',
			limit,
			prelimRate: 22,
			amountFits,
			choice: null,
		};
		item.full.locks.stage2Available = amountFits;

		return HttpResponse.json(result);
	}),
	http.post(url('/applications/:id/prelim/choice'), async ({ params, request }) => {
		const item = findApplication(String(params.id));
		const body = await request.json() as { choice: 'REDUCE_AMOUNT' | 'KEEP_AMOUNT'; amount?: number };

		if (!item) {
			return notFound();
		}

		if (body.choice === 'REDUCE_AMOUNT' && body.amount) {
			item.full.params.amount = body.amount;
			item.short.amount = body.amount;
		}

		if (item.full.prelim) {
			item.full.prelim.choice = body.choice;
		}

		item.full.stage = 2;
		item.full.locks.stage2Available = true;
		item.full.version += 1;

		return HttpResponse.json({
			stage: 2,
			params: { amount: item.full.params.amount, term: item.full.params.term },
			version: item.full.version,
		});
	}),
	http.post(url('/applications/:id/validate'), () => HttpResponse.json({ valid: true, missing: [] })),
	http.post(url('/applications/:id/submit'), ({ params }) => {
		const item = findApplication(String(params.id));

		if (!item) {
			return notFound();
		}

		const number = `КЗ-2026-${String(item.short.id.replace(/\D/g, '')).padStart(6, '0')}`;
		item.short.number = number;
		item.full.number = number;
		item.submittedAt = nowIso();
		setStatus(item, 'SUBMITTED');

		return HttpResponse.json({
			id: item.short.id,
			number,
			status: 'SUBMITTED',
			redirect: `/applications/${item.short.id}`,
		});
	}),
	http.get(url('/applications/:id/timeline'), ({ params }) => {
		const item = findApplication(String(params.id));

		return item ? HttpResponse.json(timelineOf(item)) : notFound();
	}),
	http.get(url('/applications/:id/info-requests'), ({ params }) => {
		const item = findApplication(String(params.id));

		return item ? HttpResponse.json(item.infoRequests) : notFound();
	}),
	http.post(url('/applications/:id/info-requests/:requestId/answer'), ({ params }) => {
		const item = findApplication(String(params.id));

		if (!item) {
			return notFound();
		}

		const request = item.infoRequests.find((entry) => entry.id === params.requestId);

		if (request) {
			request.status = 'ANSWERED';
		}

		setStatus(item, 'SCORING');

		return HttpResponse.json({ requestStatus: 'ANSWERED', applicationStatus: 'SCORING' });
	}),
	http.get(url('/applications/:id/offer'), ({ params }) => {
		const item = findApplication(String(params.id));

		return item?.offer ? HttpResponse.json(item.offer) : notFound();
	}),
	http.post(url('/applications/:id/offer/accept'), ({ params }) => {
		const item = findApplication(String(params.id));

		if (!item) {
			return notFound();
		}

		setStatus(item, 'SIGNING');

		return HttpResponse.json({ status: 'SIGNING', redirect: `/applications/${item.short.id}/sign` });
	}),
	http.post(url('/applications/:id/offer/decline'), ({ params }) => {
		const item = findApplication(String(params.id));

		if (item) {
			setStatus(item, 'CLOSED');
		}

		return HttpResponse.json({ status: 'CLOSED' });
	}),
	http.get(url('/applications/:id/contract'), ({ params }) => {
		const item = findApplication(String(params.id));
		const offer = item?.offer && item.offer.decision !== 'REJECTED' ? item.offer.offered : null;

		return HttpResponse.json({
			...db.contract,
			terms: offer
				? { ...db.contract.terms, ...offer, firstPaymentDate: db.contract.terms.firstPaymentDate }
				: db.contract.terms,
			signed: item ? item.short.status === 'ISSUED' || item.loanId !== null || db.contract.signed : false,
			disbursed: item?.loanId !== null,
		});
	}),
	http.post(url('/applications/:id/contract/sign'), () => {
		db.contract.signed = true;

		return HttpResponse.json({ signed: true, signedAt: nowIso(), signatureId: nextId('sig') });
	}),
	http.post(url('/applications/:id/contract/disburse'), ({ params }) => {
		const item = findApplication(String(params.id));
		const loanId = nextId('l');

		if (item) {
			const offered = item.offer && item.offer.decision !== 'REJECTED' ? item.offer.offered : null;
			const amount = offered?.amount ?? item.short.amount;
			const term = offered?.term ?? item.short.term;
			const rate = offered?.rate ?? 22;
			const payment = annuity(amount, term, rate);
			item.loanId = loanId;
			setStatus(item, 'ISSUED');
			db.loans.push({
				loan: {
					id: loanId,
					product: item.short.product,
					amount,
					balance: amount,
					rate,
					term,
					issuedAt: nowIso().slice(0, 10),
					nextPayment: { date: '2026-11-20', amount: payment },
					status: 'ACTIVE',
				},
				details: {
					id: loanId,
					applicationId: item.short.id,
					contractNumber: `КД-2026-${loanId.replace(/\D/g, '')}`,
					terms: { amount, term, rate, monthlyPayment: payment },
					payments: Array.from({ length: Math.min(term, 12) }, (_, index) => ({
						date: `2026-${String(11 + index > 12 ? index - 1 : 11 + index).padStart(2, '0')}-20`,
						principal: Math.round(amount / term),
						interest: payment - Math.round(amount / term),
						total: payment,
						status: 'PLANNED' as const,
					})),
				},
			});
		}

		return HttpResponse.json({ status: 'ISSUED', loanId, disbursedAt: nowIso(), accountMasked: '40702••••••1234' });
	}),

	// ── кредиты ────────────────────────────────────────────
	http.get(url('/loans/summary'), () => {
		const next = db.loans[0]?.loan.nextPayment ?? null;

		return HttpResponse.json({ activeLoans: db.loans.length, nextPayment: next });
	}),
	http.get(url('/loans/schedule/upcoming'), ({ request }) => {
		const limit = Number(new URL(request.url).searchParams.get('limit') ?? scheduleFixture.length);

		return HttpResponse.json(scheduleFixture.slice(0, limit));
	}),
	http.get(url('/loans/schedule'), () => HttpResponse.json(scheduleFixture)),
	http.get(url('/loans'), () => HttpResponse.json(db.loans.map((entry) => entry.loan))),
	http.get(url('/loans/:id'), ({ params }) => {
		const entry = db.loans.find((item) => item.loan.id === params.id);

		return entry ? HttpResponse.json(entry.details) : notFound();
	}),

	// ── уведомления, поддержка ─────────────────────────────
	http.get(url('/notifications'), () => HttpResponse.json({
		unreadCount: db.notifications.filter((item) => !item.read).length,
		items: db.notifications,
	})),
	http.post(url('/notifications/:id/read'), ({ params }) => {
		const item = db.notifications.find((entry) => entry.id === params.id);

		if (item) {
			item.read = true;
		}

		return new HttpResponse(null, { status: 204 });
	}),
	http.get(url('/support/faq'), ({ request }) => {
		const step = new URL(request.url).searchParams.get('step');

		return HttpResponse.json(faqFixture.filter((item) => !step || item.step === step));
	}),
	http.get(url('/support/tickets'), () => HttpResponse.json(db.tickets)),
	http.post(url('/support/tickets'), async ({ request }) => {
		const body = await request.json() as { subject: string; text: string; applicationId?: string };
		const id = nextId('t');
		const createdAt = nowIso();
		db.tickets.unshift({
			id,
			subject: body.subject,
			status: 'OPEN',
			reply: null,
			applicationId: body.applicationId ?? null,
			updatedAt: createdAt,
		});

		return HttpResponse.json({ id, status: 'OPEN', createdAt }, { status: 201 });
	}),

	// ── планер (API_SPEC_ADDITIONAL) ───────────
	// печатная форма документа к подписанию (API_SPEC 4.7)
	http.get(url('/applications/:id/contract/documents/:docId'), ({ params }) => new HttpResponse(
		`Печатная форма ${String(params.docId)} (демо)`,
		{ headers: { 'content-type': 'text/plain; charset=utf-8' } },
	)),
	...plannerHandlers,

	// ── рабочее место банка ────────────────────────────────
	http.get(url('/bank/applications/stats'), () => {
		const submitted = db.applications.filter((item) => item.submittedAt);
		const count = (status: ApplicationStatus) => submitted.filter((item) => item.short.status === status).length;

		return HttpResponse.json({
			...bankStatsFixture,
			SUBMITTED: count('SUBMITTED') + bankStatsFixture.SUBMITTED,
			ON_DECISION: count('ON_DECISION') + bankStatsFixture.ON_DECISION,
		});
	}),
	http.get(url('/bank/applications'), ({ request }) => {
		const params = new URL(request.url).searchParams;
		const status = params.get('status');
		const items = db.applications
			.filter((item) => item.submittedAt)
			.filter((item) => !status || item.short.status === status)
			.map((item) => ({
				id: item.short.id,
				number: item.short.number,
				company: { inn: '7712345678', name: 'ООО «Северная пекарня»' },
				product: item.short.product,
				amount: item.short.amount,
				term: item.short.term,
				status: item.short.status,
				score: 640,
				riskBand: 'MEDIUM' as const,
				decisionMode: 'MANUAL' as const,
				assignee: item.assignee,
				submittedAt: item.submittedAt,
			}));

		return HttpResponse.json({
			items,
			page: Number(params.get('page') ?? 1),
			size: Number(params.get('size') ?? 20),
			total: items.length,
		});
	}),
	http.get(url('/bank/applications/:id'), ({ params }) => {
		const item = findApplication(String(params.id));

		return item
			? HttpResponse.json({
				...bankDetailsFixture,
				id: item.short.id,
				number: item.short.number,
				status: item.short.status,
				request: {
					...bankDetailsFixture.request,
					productCode: item.short.product.code,
					amount: item.short.amount,
					term: item.short.term,
				},
			})
			: notFound();
	}),
	http.post(url('/bank/applications/:id/assign'), ({ params }) => {
		const item = findApplication(String(params.id));
		const assignee = { id: 'u-9', name: 'Иванова Е. П.' };

		if (item) {
			item.assignee = assignee;
		}

		return HttpResponse.json({ assignee });
	}),
	http.get(url('/bank/applications/:id/completeness'), () => HttpResponse.json(completenessFixture)),
	http.get(url('/bank/applications/:id/checks'), () => HttpResponse.json(checksFixture)),
	http.post(url('/bank/applications/:id/checks/refresh'), () => HttpResponse.json(
		checksFixture.map((check) => ({ ...check, requestedAt: nowIso() })),
	)),
	http.get(url('/bank/applications/:id/score'), () => HttpResponse.json(scoreFixture)),
	http.get(url('/bank/applications/:id/history'), () => HttpResponse.json(historyFixture)),
	http.post(url('/bank/applications/:id/info-requests'), async ({ params, request }) => {
		const item = findApplication(String(params.id));
		const body = await request.json() as { comment: string; items: { key: string; type: string; label: string }[] };
		const id = nextId('r');

		if (item) {
			item.infoRequests.push({
				id,
				status: 'OPEN',
				comment: body.comment,
				items: body.items as never,
				createdAt: nowIso(),
			});
			setStatus(item, 'INFO_REQUESTED');
		}

		return HttpResponse.json({ id, status: 'OPEN', applicationStatus: 'INFO_REQUESTED' }, { status: 201 });
	}),
	http.post(url('/bank/applications/:id/decision'), async ({ params, request }) => {
		const item = findApplication(String(params.id));
		const body = await request.json() as {
			type: 'APPROVE' | 'COUNTER' | 'REJECT';
			amount?: number;
			term?: number;
			productCode?: string;
			rate?: number;
			reason?: string;
		};

		if (!item) {
			return notFound();
		}

		const requested = { productCode: item.short.product.code, amount: item.short.amount, term: item.short.term };

		if (body.type === 'REJECT') {
			item.offer = {
				decision: 'REJECTED',
				reason: body.reason ?? 'Отказ по результатам скоринга',
				alternatives: [{ productCode: 'EXPRESS', maxAmount: 500000 }],
			} satisfies Offer;
			setStatus(item, 'REJECTED');
		} else {
			const amount = body.amount ?? requested.amount;
			const term = body.term ?? requested.term;
			const rate = body.rate ?? 22;
			const validUntil = new Date(Date.now() + 7 * 86400000).toISOString().slice(0, 10);
			item.offer = {
				decision: body.type === 'APPROVE' ? 'APPROVED' : 'COUNTER_OFFER',
				requested,
				offered: {
					productCode: body.productCode ?? requested.productCode,
					amount,
					term,
					rate,
					monthlyPayment: annuity(amount, term, rate),
				},
				explanation: body.type === 'COUNTER' ? 'Условия скорректированы по итогам скоринга' : null,
				validUntil,
			} satisfies Offer;
			setStatus(item, body.type === 'APPROVE' ? 'APPROVED' : 'COUNTER_OFFER');
		}

		return HttpResponse.json({ status: item.short.status, decisionId: nextId('dec'), mode: 'MANUAL' });
	}),
];
