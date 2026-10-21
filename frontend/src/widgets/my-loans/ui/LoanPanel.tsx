import { type Loan, type LoanPayment, PaymentStatusBadge, useGetLoanQuery } from '@entities/loan';
import { ROUTES } from '@shared/config';
import { formatDate, formatMoney, formatRate, formatTerm } from '@shared/lib/format';
import { AsyncBoundary, Badge, ButtonLink, DataTable, Icon } from '@shared/ui';
import { Link } from 'react-router';

import {
	chartCeiling,
	monthLabel,
	paymentDay,
	remainingCount,
	repaidPercent,
	thousands,
	UPCOMING_COUNT,
	upcomingPayments,
} from '../model';
import styles from './LoanPanel.module.scss';

const scheduleId = (loanId: string) => `payment-schedule-${loanId}`;

const Money = ({ value }: { value: number }) => (
	<div className={styles.largeMoney}>{value.toLocaleString('ru-RU')} <span>₽</span></div>
);

const ScheduleBadge = ({ payment, first }: { payment: LoanPayment; first: boolean }) =>
	payment.status === 'PLANNED' ? (
		<Badge tone={first ? 'warning' : 'neutral'}>{first ? 'Предстоит' : 'По графику'}</Badge>
	) : (
		<PaymentStatusBadge status={payment.status} />
	);

const Chart = ({ payments }: { payments: LoanPayment[] }) => {
	const ceiling = chartCeiling(payments);

	return (
		<div className={styles.chart}>
			<div className={styles.axis} aria-hidden="true">
				<span>{thousands(ceiling)}</span>
				<span>{thousands(ceiling / 2)}</span>
				<span>0</span>
			</div>
			<ul className={styles.bars}>
				{payments.map((payment) => (
					<li key={payment.date} className={styles.column}>
						<div
							className={styles.stack}
							role="img"
							aria-label={`${monthLabel(payment.date)}: долг ${formatMoney(payment.principal)}, проценты ${formatMoney(payment.interest)}`}
						>
							<span className={styles.interest} style={{ height: `${(payment.interest / ceiling) * 100}%` }} />
							<span className={styles.principal} style={{ height: `${(payment.principal / ceiling) * 100}%` }} />
							<span className={styles.tooltip}>
								{monthLabel(payment.date)}: {formatMoney(payment.principal)} долг<br />
								{formatMoney(payment.interest)} проценты
							</span>
						</div>
						<small>{monthLabel(payment.date)}</small>
					</li>
				))}
			</ul>
		</div>
	);
};

/** Один действующий кредит по макету loans.html: сводка, ближайший платёж, состав платежа и график */
export const LoanPanel = ({ loan }: { loan: Loan }) => {
	const { data, isLoading, isError, refetch } = useGetLoanQuery(loan.id);
	const upcoming = data ? upcomingPayments(data.payments) : [];
	const left = data ? remainingCount(data.payments) : 0;
	const next = loan.nextPayment ? paymentDay(loan.nextPayment.date) : null;
	const nextRow = upcoming[0];
	const repaid = repaidPercent(loan);

	return (
		<AsyncBoundary
			isLoading={isLoading}
			isError={isError}
			onRetry={() => void refetch()}
			loadingLabel="Загружаем договор"
			errorText="Не удалось загрузить детали кредита"
		>
			<div className={styles.panel}>
				<div className={styles.overview}>
					<section className={styles.card} aria-labelledby={`loan-${loan.id}`}>
						<div className={styles.heading}>
							<div>
								<span className={styles.eyebrow}>ДОГОВОР {data?.contractNumber}</span>
								<h2 id={`loan-${loan.id}`}>{loan.product.name}</h2>
							</div>
							<Badge tone="success">Действует</Badge>
						</div>
						<div className={styles.summaryBody}>
							<div>
								<span className={styles.muted}>Остаток основного долга</span>
								<Money value={loan.balance} />
								<dl className={styles.keyValues}>
									<div><dt>Первоначальная сумма</dt><dd>{formatMoney(loan.amount)}</dd></div>
									<div><dt>Ставка</dt><dd>{formatRate(loan.rate)}</dd></div>
									<div><dt>Срок</dt><dd>{formatTerm(loan.term)}</dd></div>
									<div><dt>До окончания</dt><dd>{left} платежей</dd></div>
								</dl>
							</div>
							<div className={styles.donutWrap}>
								<div
									className={styles.donut}
									style={{ background: `conic-gradient(var(--color-primary) 0% ${repaid}%, var(--color-raised) ${repaid}% 100%)` }}
								>
									<div>
										<strong>{repaid}<span>%</span></strong>
										<small>долга погашено</small>
									</div>
								</div>
								<span>{formatMoney(loan.amount - loan.balance)} из {formatMoney(loan.amount)}</span>
							</div>
						</div>
						<div className={styles.links}>
							<Link to={ROUTES.documents} className={styles.textLink}>
								Кредитный договор <Icon name="arrow" size={14} />
							</Link>
						</div>
					</section>
					{next && loan.nextPayment && (
						<section className={styles.focus} aria-label="Ближайший платёж">
							<span className={styles.eyebrow}>БЛИЖАЙШИЙ ПЛАТЁЖ</span>
							<h2>{next.day}</h2>
							<span className={styles.muted}>{next.year} · по графику</span>
							<Money value={loan.nextPayment.amount} />
							{nextRow && (
								<div className={styles.breakdown}>
									<span>Основной долг <strong>{formatMoney(nextRow.principal)}</strong></span>
									<span>Проценты <strong>{formatMoney(nextRow.interest)}</strong></span>
								</div>
							)}
							<a className={styles.scheduleButton} href={`#${scheduleId(loan.id)}`}>
								К графику платежей <Icon name="arrow" size={16} />
							</a>
							<p>Просроченной задолженности нет</p>
						</section>
					)}
				</div>
				<div className={styles.analysis}>
					<section className={styles.card} aria-labelledby={`chart-${loan.id}`}>
						<div className={styles.heading}>
							<div>
								<h2 id={`chart-${loan.id}`}>Из чего состоит платёж</h2>
								<p>Ближайшие {UPCOMING_COUNT} месяцев · сумма в рублях</p>
							</div>
							{nextRow && <span className={styles.total}>{formatMoney(nextRow.total)} / мес.</span>}
						</div>
						<div className={styles.legend}>
							<span><i className={styles.principalDot} />Основной долг</span>
							<span><i className={styles.interestDot} />Проценты</span>
						</div>
						<Chart payments={upcoming} />
						<p className={styles.caption}>Наведите на столбец, чтобы увидеть состав платежа.</p>
					</section>
					<section className={`${styles.card} ${styles.related}`} aria-label="Всё по вашему кредиту">
						<span className={styles.tile}>ВАЖНОЕ РЯДОМ</span>
						<h2>Всё по вашему кредиту</h2>
						<Link to={ROUTES.documents} className={styles.relatedLink}>
							Кредитный договор <Icon name="arrow" size={14} />
						</Link>
						<a href={`#${scheduleId(loan.id)}`} className={styles.relatedLink}>
							График платежей <Icon name="arrow" size={14} />
						</a>
						<Link to={ROUTES.support} className={styles.relatedLink}>
							Задать вопрос <Icon name="arrow" size={14} />
						</Link>
						<div className={styles.tip}>
							<Icon name="check" size={18} />
							<p>Следующий платёж — по графику.<br />Проверьте остаток на счёте заранее.</p>
						</div>
					</section>
				</div>
				<section className={`${styles.card} ${styles.schedule}`} id={scheduleId(loan.id)} aria-labelledby={`schedule-${loan.id}`}>
					<div className={styles.heading}>
						<div>
							<h2 id={`schedule-${loan.id}`}>График платежей</h2>
							<p>Ближайшие платежи по действующему кредиту</p>
						</div>
						<ButtonLink variant="secondary" to={ROUTES.creditsCalendar} className={styles.calendarLink}>
							Платёжный календарь
						</ButtonLink>
					</div>
					<DataTable
						caption={`Ближайшие ${upcoming.length} платежей`}
						hideCaption
						className={styles.table}
						rows={upcoming}
						getRowKey={(payment) => payment.date}
						columns={[
							{ key: 'date', header: 'Дата', render: (payment) => formatDate(payment.date) },
							{ key: 'principal', header: 'Основной долг', render: (payment) => formatMoney(payment.principal) },
							{ key: 'interest', header: 'Проценты', render: (payment) => formatMoney(payment.interest) },
							{ key: 'total', header: 'Всего', render: (payment) => <strong>{formatMoney(payment.total)}</strong> },
							{
								key: 'status',
								header: 'Статус',
								render: (payment) => <ScheduleBadge payment={payment} first={payment === nextRow} />,
							},
						]}
					/>
					<p className={styles.note}>Показаны ближайшие {upcoming.length} из {left} оставшихся платежей.</p>
				</section>
			</div>
		</AsyncBoundary>
	);
};
