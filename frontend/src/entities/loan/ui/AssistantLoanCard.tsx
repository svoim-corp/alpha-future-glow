import { formatMoney, formatTerm } from '@shared/lib/format';
import { ButtonLink, Card, KeyValueList } from '@shared/ui';

interface AssistantLoanCardProps {
	loan: {
		id: string; productName: string; amount: number; balance: number; rate: number; term: number; status: string;
	};
}

export const AssistantLoanCard = ({ loan: l }: AssistantLoanCardProps) => (
	<Card>
		<h3>{l.productName}</h3>
		<KeyValueList items={[
			{ label: 'Кредит', value: l.id },
			{ label: 'Выдано', value: formatMoney(l.amount) },
			{ label: 'Остаток основного долга', value: formatMoney(l.balance) },
			{ label: 'Ставка', value: `${l.rate} % годовых` },
			{ label: 'Срок', value: formatTerm(l.term) },
			{ label: 'Статус', value: l.status === 'ACTIVE' ? 'Действует' : 'Закрыт' },
		]} />
		<ButtonLink to="/credits" variant="secondary">Открыть мои кредиты</ButtonLink>
	</Card>
);
