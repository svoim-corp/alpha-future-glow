import { formatMoney, formatTerm } from '@shared/lib/format';
import { Card, KeyValueList } from '@shared/ui';
import { Link } from 'react-router';

interface AssistantApplicationCardProps {
	application: { id: string; number: string | null; amount: number; term: number; status: string; nextStep: string };
}

export const AssistantApplicationCard = ({ application: a }: AssistantApplicationCardProps) => (
	<Card>
		<h3>{a.number ? `Заявка № ${a.number}` : 'Черновик заявки'}</h3>
		<KeyValueList items={[
			{ label: 'Статус', value: a.status },
			{ label: 'Сумма', value: formatMoney(a.amount) },
			{ label: 'Срок', value: formatTerm(a.term) },
		]} />
		<p>{a.nextStep}</p>
		<Link to={`/applications/${encodeURIComponent(a.id)}`}>Открыть заявку</Link>
	</Card>
);
