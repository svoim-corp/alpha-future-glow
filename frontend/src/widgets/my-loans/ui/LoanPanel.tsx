import { type Loan, type LoanPayment, PaymentStatusBadge, useGetLoanQuery } from '@entities/loan';
import { paths, ROUTES } from '@shared/config';
import { formatDate, formatMoney, formatRate, formatTerm } from '@shared/lib/format';
import { Alert, AsyncBoundary, Badge, Button, ButtonLink, DataTable, Icon, SelectField } from '@shared/ui';
import { useId, useState } from 'react';
import { Link } from 'react-router';

import { chartCeiling, monthLabel, paymentDay, remainingCount, repaidPercent, scheduleCsv, thousands, upcomingPayments } from '../model';
import { LoanConditionsDialog } from './LoanConditionsDialog';
import styles from './LoanPanel.module.scss';

const scheduleId = (loanId: string) => `payment-schedule-${loanId}`;
type ScheduleMode = 'upcoming' | 'all' | 'paid' | 'overdue';

const Money = ({ value }: { value: number }) => (
	<div className={styles.largeMoney}>{value.toLocaleString('ru-RU')} <span>₽</span></div>
);

const Chart = ({ payments }: { payments: LoanPayment[] }) => {
	const ceiling = chartCeiling(payments);
	const id = useId();
	const [opened, setOpened] = useState<string | null>(null);

	return (
		<div className={styles.chart}>
			<div className={styles.axis} aria-hidden="true"><span>{thousands(ceiling)}</span><span>{thousands(ceiling / 2)}</span><span>0</span></div>
			<ul className={styles.bars}>
				{payments.map((payment) => (
					<li key={payment.date} className={styles.column}>
						<button type="button" className={styles.stack} data-open={opened === payment.date}
							aria-label={`${formatDate(payment.date)}: долг ${formatMoney(payment.principal)}, проценты ${formatMoney(payment.interest)}`}
							aria-describedby={opened === payment.date ? `${id}-${payment.date}` : undefined}
							onMouseEnter={() => setOpened(payment.date)}
							onMouseLeave={(event) => { if (document.activeElement !== event.currentTarget) setOpened(null); }}
							onFocus={() => setOpened(payment.date)} onBlur={() => setOpened(null)}
							onClick={() => setOpened(payment.date)}
							onKeyDown={(event) => { if (event.key === 'Escape') setOpened(null); }}>
							<span className={styles.interest} style={{ height: `${(payment.interest / ceiling) * 100}%` }} />
							<span className={styles.principal} style={{ height: `${(payment.principal / ceiling) * 100}%` }} />
							{opened === payment.date && <span id={`${id}-${payment.date}`} role="tooltip" className={styles.tooltip}>
								{formatDate(payment.date)}: {formatMoney(payment.principal)} долг<br />{formatMoney(payment.interest)} проценты
							</span>}
						</button>
						<small>{monthLabel(payment.date)}</small>
					</li>
				))}
			</ul>
		</div>
	);
};

/** Детали кредита и все строки серверного графика; будущие платежи отделены от просрочки. */
export const LoanPanel = ({ loan }: { loan: Loan }) => {
	const { data, isLoading, isError, refetch } = useGetLoanQuery(loan.id, { refetchOnFocus: true, refetchOnReconnect: true, refetchOnMountOrArgChange: true });
	const [conditionsOpen, setConditionsOpen] = useState(false);
	const [mode, setMode] = useState<ScheduleMode>(loan.status === 'CLOSED' ? 'all' : 'upcoming');
	const payments = data?.payments ?? [];
	const upcoming = upcomingPayments(payments);
	const overdue = payments.filter((payment) => payment.status === 'OVERDUE');
	const overdueAmount = overdue.reduce((total, payment) => total + payment.total, 0);
	const left = remainingCount(payments);
	const current = { ...loan, balance: data?.balance ?? loan.balance, status: data?.status ?? loan.status,
		amount: data?.terms.amount ?? loan.amount, rate: data?.terms.rate ?? loan.rate, term: data?.terms.term ?? loan.term };
	const closed = current.status === 'CLOSED';
	const nextRow = closed ? undefined : upcoming[0];
	const next = nextRow ? paymentDay(nextRow.date) : null;
	const repaid = repaidPercent(current);
	const rows = mode === 'all' ? payments : mode === 'paid' ? payments.filter((payment) => payment.status === 'PAID')
		: mode === 'overdue' ? overdue : upcoming;
	const download = () => {
		const url = URL.createObjectURL(new Blob([scheduleCsv(payments)], { type: 'text/csv;charset=utf-8' }));
		const link = document.createElement('a');
		link.href = url;
		link.download = `schedule-${loan.id}.csv`;
		document.body.append(link);
		link.click();
		link.remove();
		setTimeout(() => URL.revokeObjectURL(url), 1000);
	};

	return (
		<AsyncBoundary isLoading={isLoading} isError={isError} onRetry={() => void refetch()}
			loadingLabel="Загрузка договора" errorText="Не удалось загрузить детали кредита.">
			{data && <div className={styles.panel} id={`credit-${loan.id}`}>
				{overdue.length > 0 && <Alert tone="danger" title={`Просроченные платежи: ${formatMoney(overdueAmount)}`}>
					По графику не оплачено в срок платежей: {overdue.length}. Сумма не включает возможные пени.
					<a href={`#${scheduleId(loan.id)}`} onClick={() => setMode('overdue')}>Показать в графике</a>
				</Alert>}
				<div className={`${styles.overview} ${closed ? styles.single : ''}`}>
					<section className={styles.card} aria-labelledby={`loan-${loan.id}`}>
						<div className={styles.heading}>
							<div><span className={styles.eyebrow}>ДОГОВОР {data.contractNumber}</span><h2 id={`loan-${loan.id}`}>{data.product?.name ?? loan.product.name}</h2></div>
							<Badge tone={closed ? 'neutral' : 'success'}>{closed ? 'Закрыт' : 'Действует'}</Badge>
						</div>
						<div className={styles.summaryBody}>
							<div>
								<span className={styles.muted}>Остаток основного долга</span><Money value={current.balance} />
								<dl className={styles.keyValues}>
									<div><dt>Первоначальная сумма</dt><dd>{formatMoney(current.amount)}</dd></div>
									<div><dt>Ставка</dt><dd>{formatRate(current.rate)}</dd></div>
									<div><dt>Срок</dt><dd>{formatTerm(current.term)}</dd></div>
									<div><dt>Неоплаченных платежей</dt><dd>{left}</dd></div>
								</dl>
							</div>
							<div className={styles.donutWrap}>
								<div className={styles.donut} role="img" aria-label={`Погашено ${repaid}% основного долга`}
									style={{ background: `conic-gradient(var(--color-primary) 0% ${repaid}%, var(--color-raised) ${repaid}% 100%)` }}>
									<div><strong>{repaid}<span>%</span></strong><small>долга погашено</small></div>
								</div>
								<span>{formatMoney(current.amount - current.balance)} из {formatMoney(current.amount)}</span>
							</div>
						</div>
						<Button variant="secondary" onClick={() => setConditionsOpen(true)}>Условия и документы кредита</Button>
					</section>
					{!closed && <section className={styles.focus} aria-label="Ближайший платёж">
						<span className={styles.eyebrow}>БЛИЖАЙШИЙ ПЛАТЁЖ</span>
						{next && nextRow ? <>
							<h2>{next.day}</h2><span className={styles.muted}>{next.year} · по графику</span><Money value={nextRow.total} />
							<div className={styles.breakdown}>
								<span>Основной долг <strong>{formatMoney(nextRow.principal)}</strong></span>
								<span>Проценты <strong>{formatMoney(nextRow.interest)}</strong></span>
							</div>
						</> : <h2>Будущих платежей нет</h2>}
						<a className={styles.scheduleButton} href={`#${scheduleId(loan.id)}`}>К графику платежей <Icon name="arrow" size={16} /></a>
						{overdue.length === 0 && <p>По графику просроченных платежей нет</p>}
					</section>}
				</div>
				<div className={`${styles.analysis} ${closed || upcoming.length === 0 ? styles.single : ''}`}>
					{!closed && upcoming.length > 0 && <section className={styles.card} aria-labelledby={`chart-${loan.id}`}>
						<div className={styles.heading}><div><h2 id={`chart-${loan.id}`}>Из чего состоит платёж</h2><p>Ближайших платежей: {upcoming.length} · сумма в рублях</p></div></div>
						<div className={styles.legend}><span><i className={styles.principalDot} />Основной долг</span><span><i className={styles.interestDot} />Проценты</span></div>
						<Chart payments={upcoming} />
						<p className={styles.caption}>Наведите, коснитесь столбца или выберите его клавишей Tab, чтобы увидеть состав платежа.</p>
					</section>}
					<section className={`${styles.card} ${styles.related}`} aria-label="Всё по вашему кредиту">
						<span className={styles.tile}>ВАЖНОЕ РЯДОМ</span><h2>Всё по вашему кредиту</h2>
						<a href={`#${scheduleId(loan.id)}`} className={styles.relatedLink} onClick={() => setMode('all')}>Полный график платежей <Icon name="arrow" size={14} /></a>
						<Link to={paths.application(data.applicationId)} className={styles.relatedLink}>Заявка на кредит <Icon name="arrow" size={14} /></Link>
						<Link to={paths.support(data.applicationId)} className={styles.relatedLink}>Задать вопрос <Icon name="arrow" size={14} /></Link>
					</section>
				</div>
				<section className={`${styles.card} ${styles.schedule}`} id={scheduleId(loan.id)} aria-labelledby={`schedule-${loan.id}`}>
					<div className={styles.heading}>
						<div><h2 id={`schedule-${loan.id}`}>График платежей</h2><p>{closed ? 'История платежей по закрытому кредиту' : 'Будущие, оплаченные и просроченные платежи'}</p></div>
						<div className={styles.actions}><Button variant="secondary" onClick={download} disabled={payments.length === 0}>Скачать график CSV</Button>
							<ButtonLink variant="secondary" to={ROUTES.creditsCalendar}>Платёжный календарь</ButtonLink></div>
					</div>
					<SelectField label="Показать платежи" value={mode} onChange={(value) => {
						if (value === 'upcoming' || value === 'all' || value === 'paid' || value === 'overdue') setMode(value);
					}} options={[
						{ value: 'upcoming', label: 'Ближайшие будущие' }, { value: 'all', label: 'Полный график' },
						{ value: 'paid', label: 'Оплаченные' }, { value: 'overdue', label: 'Просроченные' },
					]} />
					{rows.length > 0 ? <DataTable caption="График платежей" hideCaption className={styles.table} rows={rows} getRowKey={(payment) => payment.date} columns={[
						{ key: 'date', header: 'Дата', render: (payment) => formatDate(payment.date) },
						{ key: 'principal', header: 'Основной долг', align: 'right', render: (payment) => formatMoney(payment.principal) },
						{ key: 'interest', header: 'Проценты', align: 'right', render: (payment) => formatMoney(payment.interest) },
						{ key: 'total', header: 'Всего', align: 'right', render: (payment) => <strong>{formatMoney(payment.total)}</strong> },
						{ key: 'status', header: 'Статус', render: (payment) => <PaymentStatusBadge status={payment.status} /> },
					]} /> : <p role="status" className={styles.note}>В выбранном разделе платежей нет.</p>}
					<p className={styles.note}>Показано {rows.length} из {payments.length} платежей графика.</p>
				</section>
				{conditionsOpen && <LoanConditionsDialog loan={current} details={data} onClose={() => setConditionsOpen(false)} />}
			</div>}
		</AsyncBoundary>
	);
};
