import { formatDate, formatMoney } from '@shared/lib/format';
import { Card, KeyValueList } from '@shared/ui';
import { Link } from 'react-router';

interface AssistantPaymentCardProps {
	payment: { loanId: string; productName: string; date: string; amount: number };
}

export const AssistantPaymentCard = ({ payment: p }: AssistantPaymentCardProps) => (
	<Card>
		<h3>{p.productName}</h3>
		<KeyValueList items={[
			{ label: 'Дата платежа', value: formatDate(p.date) },
			{ label: 'Сумма', value: formatMoney(p.amount) },
		]} />
		<Link to="/credits/calendar">Открыть календарь платежей</Link>
	</Card>
);
