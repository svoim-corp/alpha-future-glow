import { formatDate, formatMoney } from '@shared/lib/format';
import { ButtonLink, Card, KeyValueList } from '@shared/ui';

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
		<ButtonLink to="/credits/calendar" variant="secondary">Открыть календарь платежей</ButtonLink>
	</Card>
);
