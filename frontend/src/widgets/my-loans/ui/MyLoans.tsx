import { type ApplicationShort, useGetMyApplicationsQuery } from '@entities/application';
import { useGetLoansQuery } from '@entities/loan';
import { ROUTES } from '@shared/config';
import { formatMoney } from '@shared/lib/format';
import { AsyncBoundary, ButtonLink, Metrics, panelId, tabId, Tabs } from '@shared/ui';
import { useState } from 'react';

import { earliestPaymentDate, paymentDay, totalBalance, totalMonthly } from '../model';
import { LoanPanel } from './LoanPanel';
import styles from './MyLoans.module.scss';
import { OfferCard } from './OfferCard';

type TabKey = 'active' | 'approved' | 'closed';

const isOffer = (item: ApplicationShort) => item.status === 'APPROVED' || item.status === 'COUNTER_OFFER';

/** «Мои кредиты» по макету loans.html: обзор, вкладки «Действующие / Одобренные / Закрытые» */
export const MyLoans = () => {
	const { data, isLoading, isError, refetch } = useGetLoansQuery();
	const { data: applications } = useGetMyApplicationsQuery();
	const [tab, setTab] = useState<TabKey>('active');
	const active = data?.filter((loan) => loan.status === 'ACTIVE') ?? [];
	const closed = data?.filter((loan) => loan.status === 'CLOSED') ?? [];
	const offers = applications?.items.filter(isOffer) ?? [];
	const nextDate = earliestPaymentDate(active);
	const offerAmount = offers.reduce((total, item) => total + item.amount, 0);

	return (
		<AsyncBoundary
			isLoading={isLoading}
			isError={isError}
			isEmpty={data?.length === 0 && offers.length === 0}
			onRetry={() => void refetch()}
			loadingLabel="Загружаем кредиты"
			errorText="Не удалось загрузить кредиты"
			emptyTitle="Выданных кредитов пока нет"
			emptyText="Когда банк выдаст кредит, он появится здесь вместе с графиком платежей."
			emptyAction={<ButtonLink to={ROUTES.products}>Подобрать финансирование</ButtonLink>}
		>
			<div className={styles.root}>
				<Metrics
					label="Обзор кредитов"
					items={[
						{
							label: 'Основной долг',
							value: formatMoney(totalBalance(active)),
							hint: `Действующих кредитов: ${active.length}`,
						},
						{
							label: 'Ежемесячный платёж',
							value: formatMoney(totalMonthly(active)),
							hint: nextDate ? `Следующий — ${paymentDay(nextDate).day}` : 'Платежей нет',
							accent: true,
						},
						{
							label: 'Готовое предложение',
							value: formatMoney(offerAmount),
							hint: offers.length > 0 ? 'Ещё не выдано' : 'Одобренных предложений нет',
						},
					]}
				/>
				<Tabs
					label="Состояние кредитов"
					value={tab}
					onChange={setTab}
					items={[
						{ id: 'active', label: 'Действующие', count: active.length },
						{ id: 'approved', label: 'Одобренные', count: offers.length },
						{ id: 'closed', label: 'Закрытые', count: closed.length },
					]}
				/>
				<section id={panelId(tab)} role="tabpanel" aria-labelledby={tabId(tab)} className={styles.panel}>
					{tab === 'active' && (active.length > 0
						? active.map((loan) => <LoanPanel key={loan.id} loan={loan} />)
						: <Empty title="Действующих кредитов нет" text="Когда банк выдаст кредит, он появится здесь." />)}
					{tab === 'approved' && (offers.length > 0
						? offers.map((item) => <OfferCard key={item.id} application={item} />)
						: <Empty title="Одобренных предложений нет" text="Одобренные банком заявки появятся здесь." />)}
					{tab === 'closed' && (closed.length > 0
						? closed.map((loan) => <LoanPanel key={loan.id} loan={loan} />)
						: <Empty title="Закрытых кредитов пока нет" text="После полного погашения кредит и его документы останутся здесь." />)}
				</section>
			</div>
		</AsyncBoundary>
	);
};

const Empty = ({ title, text }: { title: string; text: string }) => (
	<div className={styles.empty}>
		<h2>{title}</h2>
		<p>{text}</p>
	</div>
);
