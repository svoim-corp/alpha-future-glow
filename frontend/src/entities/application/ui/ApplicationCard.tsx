import { formatDateTime, formatMoney, formatTerm } from '@shared/lib/format';
import { Badge, ButtonLink, Card } from '@shared/ui';
import { type ReactNode } from 'react';
import { useId } from 'react';

import { applicationHref } from '../model/status';
import { type ApplicationShort } from '../model/types';
import styles from './ApplicationCard.module.scss';
import { StatusBadge } from './StatusBadge';

interface ApplicationCardProps {
	application: ApplicationShort;
	/** Дополнительные действия рядом с основной ссылкой, например «Удалить черновик» */
	actions?: ReactNode;
}

export const ApplicationCard = ({ application, actions }: ApplicationCardProps) => {
	const titleId = useId();
	const isDraft = application.status === 'DRAFT';

	return (
		<Card as="article" className={styles.card} aria-labelledby={titleId}>
			<div className={styles.head}>
				<div>
					<h3 id={titleId} className={styles.title}>{application.product.name}</h3>
					<span className={styles.number}>{application.number ?? 'Номер появится после отправки'}</span>
				</div>
				<div className={styles.footer}>
					<StatusBadge status={application.status} />
					{application.actionRequired && <Badge tone="warning">Нужно действие</Badge>}
				</div>
			</div>
			<dl className={styles.facts}>
				<div className={styles.fact}>
					<dt>Сумма</dt>
					<dd>{formatMoney(application.amount)}</dd>
				</div>
				<div className={styles.fact}>
					<dt>Срок</dt>
					<dd>{formatTerm(application.term)}</dd>
				</div>
				<div className={`${styles.fact} ${styles.small}`}>
					<dt>Обновлена</dt>
					<dd>{formatDateTime(application.updatedAt)}</dd>
				</div>
			</dl>
			<div className={styles.footer}>
				<ButtonLink to={applicationHref(application)} variant={application.actionRequired ? 'primary' : 'secondary'} compact>
					{isDraft ? 'Продолжить заполнение' : 'Открыть заявку'}
				</ButtonLink>
				{actions}
			</div>
		</Card>
	);
};
