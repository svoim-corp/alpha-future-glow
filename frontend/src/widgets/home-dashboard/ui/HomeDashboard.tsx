import {
	applicationHref,
	type ApplicationShort,
	StatusBadge,
	useGetInfoRequestsQuery,
	useGetMyApplicationsQuery,
} from '@entities/application';
import { useGetLoanQuery, useGetLoansQuery, useGetLoansSummaryQuery } from '@entities/loan';
import { PlanRecommendations } from '@entities/plan';
import { ROUTES } from '@shared/config';
import { formatTerm } from '@shared/lib/format';
import { Icon, type IconName } from '@shared/ui';
import clsx from 'clsx';
import { type ReactNode } from 'react';
import { Link } from 'react-router';

import { dateParts, formatNumber, isInProgress, TRACK, trackStage } from '../model';
import styles from './HomeDashboard.module.scss';

const Arrow = ({ size = 15 }: { size?: number }) => <Icon name="arrow" size={size} />;

const Tile = ({ icon, to, children }: { icon: IconName; to: string; children: ReactNode }) => (
	<Link to={to} className={styles.quickItem}>
		<span><Icon name={icon} size={18} /></span>
		{children}
	</Link>
);

const Money = ({ value }: { value: number }) => <>{formatNumber(value)} <span>₽</span></>;

const ApplicationsTile = ({ items }: { items: ApplicationShort[] }) => {
	const current = items.find((item) => item.actionRequired && isInProgress(item)) ?? items.find(isInProgress);
	const draft = items.find((item) => item.status === 'DRAFT');
	const requests = useGetInfoRequestsQuery(current?.id ?? '', { skip: current?.status !== 'INFO_REQUESTED' });
	const open = requests.data?.find((item) => item.status === 'OPEN');
	const stage = current ? trackStage(current.status) : 0;

	return (
		<section className={clsx(styles.widget, styles.applications)} aria-labelledby="home-applications">
			<div className={styles.heading}>
				<h2 id="home-applications">Мои заявки <span className={styles.count}>{items.length}</span></h2>
				<Link to={ROUTES.applications} className={styles.textLink}>Все заявки <Arrow /></Link>
			</div>
			{current ? (
				<>
					<div className={styles.appMain}>
						<div>
							<span className={styles.overline}>{current.product.name.toUpperCase()}</span>
							<div className={styles.amount}><Money value={current.amount} /></div>
							<span className={styles.fine}>
								{formatTerm(current.term)}{current.number ? ` · Заявка ${current.number}` : ''}
							</span>
						</div>
						<StatusBadge status={current.status} />
					</div>
					<ol className={styles.track} aria-label="Этап рассмотрения">
						{TRACK.map((label, index) => {
							const done = stage > index + 1;
							const isCurrent = stage === index + 1;

							return (
								<li
									key={label}
									className={clsx(done && styles.done, isCurrent && styles.current)}
									aria-current={isCurrent ? 'step' : undefined}
								>
									<i>{done ? <Icon name="check" size={13} /> : index + 1}</i>
									{label}
								</li>
							);
						})}
					</ol>
					{current.actionRequired && (
						<div className={styles.request}>
							<span className={styles.requestIcon}><Icon name="documents" /></span>
							<div>
								<strong>{open?.items[0]?.label ?? 'Нужно ваше действие'}</strong>
								<p>{open ? 'Банк продолжит рассмотрение после вашего ответа.' : 'Откройте заявку, чтобы продолжить.'}</p>
							</div>
							<Link to={applicationHref(current)} className={styles.requestButton}>
								{open ? 'Дополнить' : 'Открыть'} <Arrow />
							</Link>
						</div>
					)}
				</>
			) : (
				<p className={styles.empty}>Отправленных заявок пока нет. Подберите продукт в каталоге.</p>
			)}
			{draft && (
				<Link to={applicationHref(draft)} className={styles.draft}>
					<span className={styles.draftIcon}><Icon name="applications" size={18} /></span>
					<span>
						<strong>{draft.product.name}</strong>
						<small>Черновик · {formatNumber(draft.amount)} ₽ · {formatTerm(draft.term)}</small>
					</span>
					<span className={styles.draftAction}>Продолжить <Arrow size={14} /></span>
				</Link>
			)}
		</section>
	);
};

const PaymentTile = () => {
	const summary = useGetLoansSummaryQuery();
	const loans = useGetLoansQuery();
	const loan = loans.data?.find((item) => item.status === 'ACTIVE');
	const details = useGetLoanQuery(loan?.id ?? '', { skip: !loan });
	const next = summary.data?.nextPayment;
	const paid = details.data?.payments.filter((payment) => payment.status === 'PAID').length ?? 0;
	const total = details.data?.terms.term ?? 0;
	const percent = total > 0 ? Math.round((paid / total) * 100) : 0;
	const parts = next ? dateParts(next.date) : null;

	return (
		<section className={clsx(styles.widget, styles.payment)} aria-labelledby="home-payment">
			<div className={styles.heading}>
				<h2 id="home-payment">Ближайший платёж</h2>
				<span className={styles.tileIcon}><Icon name="calendar" size={19} /></span>
			</div>
			{next && parts ? (
				<>
					<div className={styles.paymentDate}>
						<span>{parts.day}</span>
						<div>{parts.month}<small>{parts.year} · {parts.weekday}</small></div>
					</div>
					<div className={styles.amount}><Money value={next.amount} /></div>
					<p className={styles.fine}>По действующему кредиту{loan ? ` «${loan.product.name}»` : ''}</p>
					{total > 0 && (
						<div className={styles.progress}>
							<div className={styles.progressLabel}>
								<span>Погашено {paid} из {total} платежей</span>
								<strong>{percent}%</strong>
							</div>
							<div
								className={styles.progressTrack}
								role="progressbar"
								aria-label="Выполнено платежей"
								aria-valuenow={percent}
								aria-valuemin={0}
								aria-valuemax={100}
							>
								<span style={{ width: `${percent}%` }} />
							</div>
						</div>
					)}
					<Link to={ROUTES.credits} className={styles.bottomLink}>Посмотреть график <Arrow size={16} /></Link>
				</>
			) : (
				<>
					<p className={styles.empty}>Действующих кредитов пока нет — платежей не предстоит.</p>
					<Link to={ROUTES.products} className={clsx(styles.bottomLink, styles.pushDown)}>
						Подобрать финансирование <Arrow size={16} />
					</Link>
				</>
			)}
		</section>
	);
};

const QuickTile = () => (
	<section className={clsx(styles.widget, styles.quick)} aria-labelledby="home-quick">
		<div className={styles.heading}>
			<h2 id="home-quick">Быстрые действия</h2>
			<span className={styles.hint}>Под рукой</span>
		</div>
		<div className={styles.quickGrid}>
			<Tile icon="catalog" to={ROUTES.products}>Подать заявку</Tile>
			<Tile icon="loans" to={ROUTES.credits}>Мои кредиты</Tile>
			<Tile icon="calendar" to={ROUTES.planner}>Открыть планер</Tile>
			<Tile icon="support" to={ROUTES.support}>Связаться с банком</Tile>
		</div>
	</section>
);

const FocusTile = ({ items }: { items: ApplicationShort[] }) => {
	const summary = useGetLoansSummaryQuery();
	const loans = useGetLoansQuery();
	const task = items.find((item) => item.actionRequired && isInProgress(item));
	const next = summary.data?.nextPayment;
	const parts = next ? dateParts(next.date) : null;
	const loan = loans.data?.find((item) => item.status === 'ACTIVE');

	return (
		<section className={clsx(styles.widget, styles.focus)} aria-labelledby="home-focus">
			<div className={styles.heading}>
				<h2 id="home-focus">В фокусе</h2>
				<Link to={ROUTES.planner} className={styles.iconLink} aria-label="Открыть планер"><Arrow size={17} /></Link>
			</div>
			<p className={styles.subtitle}>Важные задачи и ближайшие даты</p>
			{task && (
				<Link to={applicationHref(task)} className={styles.focusItem}>
					<span className={clsx(styles.marker, styles.warning)}>!</span>
					<div>
						<span className={styles.taskLabel}>ТРЕБУЕТ ВНИМАНИЯ</span>
						<strong>{task.status === 'INFO_REQUESTED' ? 'Ответить на запрос банка' : 'Открыть заявку'}</strong>
						<small>{task.product.name}{task.number ? ` · ${task.number}` : ''}</small>
					</div>
					<Icon name="chevron" size={15} />
				</Link>
			)}
			{next && parts && (
				<Link to={ROUTES.credits} className={styles.focusItem}>
					<span className={styles.marker}>{parts.day}<small>{parts.shortMonth}</small></span>
					<div>
						<span className={styles.taskLabel}>ПРЕДСТОИТ</span>
						<strong>Платёж по кредиту</strong>
						<small>{formatNumber(next.amount)} ₽{loan ? ` · ${loan.product.name}` : ''}</small>
					</div>
					<Icon name="chevron" size={15} />
				</Link>
			)}
			{!task && !next && <p className={styles.empty}>Срочных задач нет.</p>}
		</section>
	);
};

const FundingTile = () => (
	<section className={clsx(styles.widget, styles.funding)} aria-labelledby="home-funding">
		<div className={styles.art} aria-hidden="true"><div /><div /><div /></div>
		<div className={styles.fundingContent}>
			<span className={styles.fundingEyebrow}>ВОЗМОЖНОСТИ ДЛЯ БИЗНЕСА</span>
			<h2 id="home-funding">Большие планы?<br />Найдём решение.</h2>
			<p>На оборотные средства,<br />оборудование и новые идеи.</p>
			<Link to={ROUTES.products}>Подобрать финансирование <Arrow size={18} /></Link>
		</div>
	</section>
);

/** Главная по макету home.html: плитки заявок, платежа, быстрых действий, задач, каталога и планов */
export const HomeDashboard = () => {
	const { data } = useGetMyApplicationsQuery();
	const items = data?.items ?? [];

	return (
		<div className={styles.grid}>
			<ApplicationsTile items={items} />
			<PaymentTile />
			<QuickTile />
			<FocusTile items={items} />
			<FundingTile />
			<PlanRecommendations className={styles.full} />
		</div>
	);
};
