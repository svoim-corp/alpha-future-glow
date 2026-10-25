import { formatMoney, formatMoneyRange, formatRateRange, formatTerm } from '@shared/lib/format';
import { Card, KeyValueList } from '@shared/ui';

interface AssistantCalculationCardProps {
	calculation: { amount: number; term: number; min: number; max: number; rateMin: number; rateMax: number };
}

export const AssistantCalculationCard = ({ calculation: c }: AssistantCalculationCardProps) => (
	<Card>
		<h3>Предварительный платёж</h3>
		<KeyValueList items={[
			{ label: 'Сумма', value: formatMoney(c.amount) },
			{ label: 'Срок', value: formatTerm(c.term) },
			{ label: 'Ставка', value: formatRateRange(c.rateMin, c.rateMax) },
			{ label: 'Платёж в месяц', value: formatMoneyRange(c.min, c.max) },
		]} />
		<p>Расчёт ориентировочный. Индивидуальные условия определяются после проверки заявки.</p>
	</Card>
);
