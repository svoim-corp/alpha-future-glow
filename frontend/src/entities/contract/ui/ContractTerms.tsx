import { formatDate, formatMoney, formatRate, formatTerm } from '@shared/lib/format';
import { KeyValueList } from '@shared/ui';

import { type Contract } from '../model/types';

/** Итоговые условия и документы к подписанию */
export const ContractTermsView = ({ contract }: { contract: Contract }) => {
	const { terms, documents } = contract;

	return (
		<div>
			<KeyValueList
				items={[
					{ label: 'Сумма', value: formatMoney(terms.amount) },
					{ label: 'Срок', value: formatTerm(terms.term) },
					{ label: 'Ставка', value: formatRate(terms.rate) },
					{ label: 'Ежемесячный платёж', value: formatMoney(terms.monthlyPayment) },
					{ label: 'Первый платёж', value: formatDate(terms.firstPaymentDate) },
				]}
			/>
			<h3>Документы</h3>
			<ul aria-label="Документы к подписанию">
				{documents.map((document) => (
					<li key={document.id}>
						<a href={document.url} target="_blank" rel="noreferrer">{document.title}</a>
					</li>
				))}
			</ul>
		</div>
	);
};
