import { formatDate, formatMoney, formatRate, formatTerm } from '@shared/lib/format';
import { Icon, KeyValueList } from '@shared/ui';

import { type Contract } from '../model/types';
import styles from './ContractTerms.module.scss';

/** Итоговые условия и документы к подписанию */
export const ContractTermsView = ({ contract }: { contract: Contract }) => {
	const { terms, documents } = contract;

	return (
		<div className={styles.terms}>
			<KeyValueList
				items={[
					{ label: 'Сумма', value: formatMoney(terms.amount) },
					{ label: 'Срок', value: formatTerm(terms.term) },
					{ label: 'Ставка', value: formatRate(terms.rate) },
					{ label: 'Ежемесячный платёж', value: formatMoney(terms.monthlyPayment) },
					{ label: 'Первый платёж', value: formatDate(terms.firstPaymentDate) },
				]}
			/>
			<div>
				<h3 className={styles.heading}>Документы к подписанию</h3>
				<ul className={styles.documents} aria-label="Документы к подписанию">
					{documents.map((document) => (
						<li key={document.id}>
							<a href={document.url} target="_blank" rel="noreferrer" className={styles.document}>
								<span className={styles.icon}><Icon name="documents" size={18} /></span>
								<span>{document.title}</span>
								<Icon name="arrow" size={14} />
							</a>
						</li>
					))}
				</ul>
			</div>
		</div>
	);
};
