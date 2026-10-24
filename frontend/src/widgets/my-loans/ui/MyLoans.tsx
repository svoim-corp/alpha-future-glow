import { useGetPendingOffersQuery } from '@entities/application';
import { useGetLoansQuery, useGetLoansSummaryQuery } from '@entities/loan';
import { useGetProductsQuery } from '@entities/product';
import { baseApi } from '@shared/api';
import { ROUTES } from '@shared/config';
import { formatDate, formatMoney } from '@shared/lib/format';
import { useAppDispatch } from '@shared/lib/store';
import { Alert, AsyncBoundary, Button, ButtonLink, Metrics, panelId, tabId, Tabs } from '@shared/ui';
import { useId, useState } from 'react';
import { useSearchParams } from 'react-router';

import { pendingOfferTotal } from '../model/loanPresentation';
import { LoanPanel } from './LoanPanel';
import { LoanSelector } from './LoanSelector';
import styles from './MyLoans.module.scss';
import { OfferCard } from './OfferCard';

type TabKey = 'active' | 'approved' | 'closed';
interface LoansViewProps {
	requestedLoan: string | null;
	requestedTab: string | null;
}
const freshness = { refetchOnFocus: true, refetchOnReconnect: true, refetchOnMountOrArgChange: true };

/** Обязательства и предложения банка загружаются независимо, без подстановки отсутствующих сумм. */
const LoansView = ({ requestedLoan, requestedTab }: LoansViewProps) => {
	const dispatch = useAppDispatch();
	const loans = useGetLoansQuery(undefined, freshness);
	const summary = useGetLoansSummaryQuery(undefined, freshness);
	const offersQuery = useGetPendingOffersQuery(undefined, freshness);
	const [selectedTab, setTab] = useState<TabKey | null>(null);
	const tab: TabKey = selectedTab ?? (loans.data?.some((loan) => loan.id === requestedLoan && loan.status === 'CLOSED') || requestedTab === 'closed' ? 'closed' : requestedTab === 'approved' ? 'approved' : 'active');
	const [selectedIds, setSelectedIds] = useState<Partial<Record<TabKey, string>>>({});
	const detailsId = useId();
	const products = useGetProductsQuery(undefined, { skip: tab !== 'approved' });
	const active = loans.data?.filter((loan) => loan.status === 'ACTIVE') ?? [];
	const closed = loans.data?.filter((loan) => loan.status === 'CLOSED') ?? [];
	const offers = offersQuery.data ?? [];
	const visibleLoans = tab === 'closed' ? closed : active;
	const desiredLoan = selectedIds[tab] ?? (selectedTab === null ? requestedLoan : null);
	const selectedLoan = desiredLoan
		? visibleLoans.find((loan) => loan.id === desiredLoan) ?? (selectedIds[tab] ? visibleLoans[0] : undefined)
		: visibleLoans[0];
	const selectedOffer = offers.find((entry) => entry.application.id === selectedIds.approved) ?? offers[0];
	const selectCredit = (id: string) => setSelectedIds((current) => ({ ...current, [tab]: id }));
	const today = new Date().toISOString().slice(0, 10);
	const nextDate = summary.data?.nextPayment?.date;
	const loansKnown = loans.data !== undefined && !loans.isError;
	const totalsKnown = summary.data !== undefined && !summary.isError;
	const offersKnown = offersQuery.data !== undefined && !offersQuery.isError;
	const amountsKnown = offersKnown && offers.every((entry) => entry.offer !== null);
	const offerAmount = pendingOfferTotal(offers, today);

	return (
		<div className={styles.root}>
			<div className={styles.toolbar}>
				<ButtonLink variant="secondary" to={ROUTES.creditsCalendar}>Платёжный календарь</ButtonLink>
				<Button variant="ghost" disabled={loans.isFetching || summary.isFetching || offersQuery.isFetching}
					onClick={() => { dispatch(baseApi.util.invalidateTags(['Loan'])); void offersQuery.refetch(); }}>
					Обновить данные
				</Button>
			</div>
			{requestedLoan && loans.data && !loans.data.some((loan) => loan.id === requestedLoan) && <Alert tone="warning" title="Запрошенный кредит не найден в загруженном списке.">
				Проверьте доступ к договору и обновите список кредитов.
			</Alert>}
			{summary.isError && <Alert tone="danger" title="Не удалось загрузить сводку кредитов."
				actions={<Button variant="secondary" onClick={() => void summary.refetch()}>Повторить</Button>} />}
			{totalsKnown && summary.data?.overdue && summary.data.overdue.amount > 0 && <Alert tone="danger" title={`Просроченная задолженность: ${formatMoney(summary.data.overdue.amount)}`}>
				Кредитов с просрочкой: {summary.data.overdue.loans}.
				Сумма включает пени; подробности — в задолженности выбранного кредита.
			</Alert>}
			<Metrics label="Обзор кредитов" items={[
				{
					label: 'Основной долг',
					value: totalsKnown && summary.data?.totalPrincipalBalance !== undefined ? formatMoney(summary.data.totalPrincipalBalance) : '—',
					hint: totalsKnown ? `Действующих кредитов: ${summary.data?.activeLoans}` : 'Сводка кредитов не загружена',
				},
				{
					label: 'Ближайший платёж',
					value: totalsKnown ? formatMoney(summary.data?.nextPayment?.amount ?? 0) : '—',
					hint: totalsKnown ? (nextDate ? `Сумма к оплате ${formatDate(nextDate)}` : 'Будущих платежей нет; просрочка указана отдельно') : 'Сводка кредитов не загружена',
					accent: true,
				},
				{
					label: 'Одобрено, ещё не выдано',
					value: amountsKnown ? formatMoney(offerAmount) : '—',
					hint: amountsKnown ? 'По условиям банка, без истёкших предложений' : 'Дождитесь загрузки условий или повторите запрос во вкладке «Одобренные»',
				},
			]} />
			<Tabs label="Состояние кредитов" value={tab} onChange={setTab} items={[
				{ id: 'active', label: 'Действующие', count: loansKnown ? active.length : undefined,
				},
				{ id: 'approved', label: 'Одобренные', count: offersKnown ? offers.length : undefined },
				{ id: 'closed', label: 'Закрытые', count: loansKnown ? closed.length : undefined,
				},
			]} />
			<section id={panelId(tab)} role="tabpanel" aria-labelledby={tabId(tab)} className={styles.panel}>
				{tab === 'approved' ? (
					<AsyncBoundary isLoading={offersQuery.isLoading} isError={offersQuery.isError}
						onRetry={() => void offersQuery.refetch()} errorText="Не удалось загрузить предложения банка."
						loadingLabel="Загрузка предложений банка" isEmpty={offers.length === 0}
						emptyTitle="Одобренных предложений нет" emptyText="Одобренные банком заявки появятся здесь.">
						{offersQuery.isFetching && <p role="status">Обновляем предложения банка…</p>}
						{selectedOffer && <>
							<LoanSelector value={selectedOffer.application.id} onChange={selectCredit}
								detailsId={detailsId}
								items={offers.map(({ application, offer }) => ({
									id: application.id,
									title: products.data?.find((product) =>
										product.code === offer?.offered.productCode)?.name
										?? (!offer || offer.offered.productCode === application.product.code ? application.product.name : 'Предложение банка'),
									caption: `Заявка ${application.number ?? application.id}`,
									amount: offer?.offered.amount,
								}))} />
							<div id={detailsId} role="region" aria-label="Выбранное предложение банка">
								<OfferCard key={selectedOffer.application.id} entry={selectedOffer} today={today}
									onRetry={() => void offersQuery.refetch()} refreshing={offersQuery.isFetching} />
							</div>
						</>}
					</AsyncBoundary>
				) : (
					<AsyncBoundary isLoading={loans.isLoading} isError={loans.isError}
						onRetry={() => void loans.refetch()} errorText="Не удалось загрузить кредиты."
						loadingLabel="Загрузка кредитов" isEmpty={visibleLoans.length === 0}
						emptyTitle={tab === 'active' ? 'Действующих кредитов нет' : 'Закрытых кредитов пока нет'}
						emptyText={tab === 'active' ? 'После выдачи кредит появится здесь вместе с графиком платежей.' : 'Закрытые кредиты появятся здесь, когда сервер передаст соответствующий статус.'}
						emptyAction={tab === 'active' ? <ButtonLink to={ROUTES.products}>Подобрать финансирование</ButtonLink> : undefined}>
						{loans.isFetching && <p role="status">Обновляем данные кредитов…</p>}
						{selectedLoan && <>
							<LoanSelector value={selectedLoan.id} onChange={selectCredit} detailsId={detailsId}
								items={visibleLoans.map((loan) => ({
									id: loan.id,
									title: loan.product.name,
									caption: `${loan.id} · от ${formatDate(loan.issuedAt)}`,
									amount: loan.amount,
								}))} />
							<div id={detailsId} role="region" aria-label="Выбранный кредит">
								<LoanPanel key={selectedLoan.id} loan={selectedLoan} />
							</div>
						</>}
					</AsyncBoundary>
				)}
			</section>
		</div>
	);
};

export const MyLoans = () => {
	const [params] = useSearchParams();
	return <LoansView key={params.toString()} requestedLoan={params.get('loan')} requestedTab={params.get('tab')} />;
};
