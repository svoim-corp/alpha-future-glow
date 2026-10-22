import { formatMoneyRange, formatRateRange } from '@shared/lib/format';
import { ButtonLink, Card, KeyValueList } from '@shared/ui';

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
		<ButtonLink to="/products" variant="secondary">Посмотреть условия в каталоге</ButtonLink>
	</Card>
);
