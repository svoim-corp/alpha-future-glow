import { formatMoney } from '@shared/lib/format';
import { Card, KeyValueList } from '@shared/ui';

interface AssistantCompanyCardProps {
	company: {
		name: string; inn: string; regAddress: string; actualAddress: string; monthlyIncome: number;
		viewerName?: string; viewerPosition?: string;
	};
}

export const AssistantCompanyCard = ({ company: c }: AssistantCompanyCardProps) => (
	<Card>
		<h3>{c.name}</h3>
		<KeyValueList items={[
			...(c.viewerName ? [{ label: 'Вы', value: c.viewerName }] : []),
			...(c.viewerPosition ? [{ label: 'Должность', value: c.viewerPosition }] : []),
			{ label: 'ИНН', value: c.inn },
			{ label: 'Юридический адрес', value: c.regAddress },
			{ label: 'Фактический адрес', value: c.actualAddress },
			{ label: 'Среднемесячные поступления', value: formatMoney(c.monthlyIncome) },
		]} />
	</Card>
);
