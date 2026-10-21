import { formatDate, formatDateTime, formatMoney, formatTerm } from '@shared/lib/format';
import { Alert, Badge, KeyValueList, type Tone } from '@shared/ui';

import {
	type BankApplicationDetails,
	type Completeness,
	type ExternalCheck,
	type HistoryEntry,
	type Score,
} from '../model/bank';
import styles from './BankReviewViews.module.scss';
import { RiskBadge } from './RiskBadge';

const SOURCE_LABEL: Record<string, string> = {
	EGRUL: 'ЕГРЮЛ',
	EGRIP: 'ЕГРИП',
	BKI: 'Бюро кредитных историй',
	TRANSPARENT_BUSINESS: 'Прозрачный бизнес',
	FNS: 'Налоговая задолженность',
	COURTS: 'Суды и исполнительные производства',
	GIRBO: 'ГИР БО',
	MCHD: 'Реестр МЧД',
	CORE_BANKING: 'Внутренние обороты',
};

const CHECK_VIEW: Record<ExternalCheck['status'], { label: string; tone: Tone }> = {
	OK: { label: 'Ок', tone: 'success' },
	WARN: { label: 'Внимание', tone: 'warning' },
	FAIL: { label: 'Проблема', tone: 'danger' },
	UNAVAILABLE: { label: 'Недоступно', tone: 'neutral' },
};

const SEVERITY: Record<Completeness['issues'][number]['severity'], { label: string; tone: Tone }> = {
	INFO: { label: 'Справка', tone: 'info' },
	WARN: { label: 'Внимание', tone: 'warning' },
	BLOCKER: { label: 'Блокирует решение', tone: 'danger' },
};

export const ApplicantView = ({ data }: { data: BankApplicationDetails }) => (
	<KeyValueList
		items={[
			{ label: 'Заявитель', value: data.applicant.fullName },
			{
				label: 'Полномочия',
				value: data.applicant.authority === 'MCHD'
					? `Доверенность до ${data.applicant.mchdValidUntil ? formatDate(data.applicant.mchdValidUntil) : '—'}`
					: 'Руководитель',
			},
			{ label: 'Компания', value: `${data.company.name}, ИНН ${data.company.inn}` },
			{ label: 'Дата регистрации', value: formatDate(data.company.regDate) },
			{ label: 'ОКВЭД', value: data.company.okved },
			{ label: 'Налоговый режим', value: data.company.taxRegime },
			{ label: 'Фактический адрес', value: data.company.actualAddress ?? 'Совпадает с адресом регистрации' },
			{
				label: 'Запрос',
				value: `${formatMoney(data.request.amount)}, ${formatTerm(data.request.term)}, ${data.request.productCode}`,
			},
			{ label: 'Цель', value: data.request.purposeText ?? data.request.goal ?? '—' },
			...(data.turnover
				? [{ label: 'Средние поступления', value: `${formatMoney(data.turnover.avgMonthlyIncome)} в месяц` }]
				: []),
			...(data.girbo
				? [{
					label: `ГИР БО за ${data.girbo.year}`,
					value: `выручка ${formatMoney(data.girbo.revenue)}, прибыль ${formatMoney(data.girbo.netProfit)}`,
				}]
				: []),
			...(data.guarantor
				? [{
					label: 'Поручитель',
					value: `${data.guarantor.fullName}, ИНН ${data.guarantor.inn}${data.guarantor.bkiConsent ? '' : ' — нет согласия на БКИ'}`,
				}]
				: []),
			...(data.prelim
				? [{
					label: 'Предварительный лимит',
					value: `${formatMoney(data.prelim.limit)}${data.prelim.amountFits ? '' : ', запрошенная сумма выше'}`,
				}]
				: []),
		]}
	/>
);

export const CompletenessView = ({ data }: { data: Completeness }) => (
	<div>
		{data.complete
			? <Alert tone="success" title="Анкета заполнена полностью" />
			: <Alert tone="warning" title="В анкете есть пробелы" />}
		<ul className={styles.list} aria-label="Замечания по анкете">
			{data.issues.map((issue) => (
				<li key={issue.code} className={styles.row}>
					<span>{issue.text}</span>
					<Badge tone={SEVERITY[issue.severity].tone}>{SEVERITY[issue.severity].label}</Badge>
				</li>
			))}
		</ul>
	</div>
);

export const ChecksView = ({ checks }: { checks: ExternalCheck[] }) => (
	<ul className={styles.list} aria-label="Внешние проверки">
		{checks.map((check) => (
			<li key={check.source} className={styles.row}>
				<span>
					<strong>{SOURCE_LABEL[check.source] ?? check.source}</strong>
					{`: ${check.summary} `}
					<span className={styles.muted}>{`запрос ${formatDateTime(check.requestedAt)}`}</span>
				</span>
				<Badge tone={CHECK_VIEW[check.status].tone}>{CHECK_VIEW[check.status].label}</Badge>
			</li>
		))}
	</ul>
);

export const ScoreView = ({ score }: { score: Score }) => (
	<div>
		<p className={styles.big}>{score.score}</p>
		<RiskBadge band={score.riskBand} />
		<p className={styles.muted}>
			{score.autoDecisionEligible
				? 'Подходит для автоматического решения'
				: `Нужен ручной разбор: ${score.autoDecisionReason ?? 'причина не указана'}`}
		</p>
		<ul className={styles.list} aria-label="Вклад факторов в оценку">
			{score.factors.map((factor) => (
				<li key={factor.code} className={styles.row}>
					<span>{factor.label}</span>
					<span className={factor.impact >= 0 ? styles.plus : styles.minus}>
						{`${factor.impact > 0 ? '+' : ''}${factor.impact} баллов`}
					</span>
				</li>
			))}
		</ul>
		{score.suggested && (
			<p>
				{`Модель предлагает: ${{ APPROVE: 'одобрить', COUNTER_OFFER: 'иные условия', REJECT: 'отказать' }[score.suggested.decision]}`}
				{score.suggested.amount ? `, ${formatMoney(score.suggested.amount)}` : ''}
				{score.suggested.rate ? `, ${score.suggested.rate} %` : ''}
			</p>
		)}
		<p className={styles.muted}>{`Версия модели: ${score.modelVersion}`}</p>
	</div>
);

const ACTOR = { CLIENT: 'Клиент', SYSTEM: 'Система', OFFICER: 'Сотрудник' } as const;

export const HistoryView = ({ history }: { history: HistoryEntry[] }) => (
	<ol className={styles.list} aria-label="История заявки">
		{history.map((entry) => (
			<li key={`${entry.at}-${entry.action}`}>
				<span className={styles.row}>
					<span>
						{`${ACTOR[entry.actor.type]}${entry.actor.name ? ` ${entry.actor.name}` : ''}: ${entry.action}`}
					</span>
					<span className={styles.muted}>{formatDateTime(entry.at)}</span>
				</span>
				{entry.comment && <span className={styles.muted}>{entry.comment}</span>}
			</li>
		))}
	</ol>
);
