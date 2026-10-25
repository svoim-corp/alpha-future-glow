import { formatMoney, formatTerm } from '@shared/lib/format';
import { ButtonLink, Card, KeyValueList } from '@shared/ui';
import { type ReactNode } from 'react';

import styles from './AssistantApplicationCard.module.scss';

interface AssistantApplicationCardProps {
	application: { id: string; number: string | null; amount: number; term: number; status: string; nextStep: string };
	actions?: ReactNode;
}

export const AssistantApplicationCard = ({ application: a, actions }: AssistantApplicationCardProps) => (
	<Card>
		<h3>{a.number ? `Заявка № ${a.number}` : 'Черновик заявки'}</h3>
		<KeyValueList items={[
			{ label: 'Статус', value: a.status },
			{ label: 'Сумма', value: formatMoney(a.amount) },
			{ label: 'Срок', value: formatTerm(a.term) },
		]} />
		<p>{a.nextStep}</p>
		<div className={styles.actions}>
			<ButtonLink to={`/applications/${encodeURIComponent(a.id)}`}>Открыть заявку</ButtonLink>
			{actions}
		</div>
	</Card>
);
