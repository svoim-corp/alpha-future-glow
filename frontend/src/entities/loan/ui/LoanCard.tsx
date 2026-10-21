import { formatDate, formatMoney, formatRate, formatTerm } from '@shared/lib/format';
import { Badge, Card } from '@shared/ui';
import { type ReactNode } from 'react';

import { type Loan } from '../model/types';
import styles from './LoanCard.module.scss';

interface LoanCardProps {
	loan: Loan;
	actions?: ReactNode;
}

export const LoanCard = ({ loan, actions }: LoanCardProps) => (
	<Card as="article" className={styles.loan} aria-labelledby={`loan-${loan.id}`}>
		<div className={styles.head}>
			<h2 id={`loan-${loan.id}`} className={styles.title}>{loan.product.name}</h2>
			<Badge tone={loan.status === 'ACTIVE' ? 'success' : 'neutral'}>
				{loan.status === 'ACTIVE' ? 'Действует' : 'Закрыт'}
			</Badge>
		</div>
		<p className={styles.balance}>{formatMoney(loan.balance)}</p>
		<p className={styles.muted}>
			{`Остаток долга · выдан ${formatMoney(loan.amount)} ${formatDate(loan.issuedAt)} · `}
			{`${formatTerm(loan.term)} · ${formatRate(loan.rate)}`}
		</p>
		{loan.nextPayment && (
			<p>{`Ближайший платёж ${formatDate(loan.nextPayment.date)} — ${formatMoney(loan.nextPayment.amount)}`}</p>
		)}
		{actions}
	</Card>
);
