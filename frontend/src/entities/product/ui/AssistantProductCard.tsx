import { formatMoneyRange, formatRateRange } from '@shared/lib/format';
import { Card, KeyValueList } from '@shared/ui';
import { Link } from 'react-router';

interface AssistantProductCardProps {
	product: {
		name: string; amountMin: number; amountMax: number;
		termMin: number; termMax: number; rateMin: number; rateMax: number;
	};
}

export const AssistantProductCard = ({ product: p }: AssistantProductCardProps) => (
	<Card>
		<h3>{p.name}</h3>
		<KeyValueList items={[
			{ label: 'Сумма', value: formatMoneyRange(p.amountMin, p.amountMax) },
			{ label: 'Срок', value: `${p.termMin}–${p.termMax} мес.` },
			{ label: 'Ставка', value: formatRateRange(p.rateMin, p.rateMax) },
		]} />
		<Link to="/products">Посмотреть условия в каталоге</Link>
	</Card>
);
