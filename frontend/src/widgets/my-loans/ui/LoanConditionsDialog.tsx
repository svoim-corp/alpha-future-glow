import { useGetContractQuery } from '@entities/contract';
import { type Loan, type LoanDetails } from '@entities/loan';
import { formatDate, formatMoney, formatRate, formatTerm } from '@shared/lib/format';
import { AsyncBoundary, Icon, KeyValueList, Modal } from '@shared/ui';

import styles from './LoanConditionsDialog.module.scss';

interface LoanConditionsDialogProps {
	loan: Loan;
	details: LoanDetails;
	onClose: () => void;
}

export const LoanConditionsDialog = ({ loan, details, onClose }: LoanConditionsDialogProps) => {
	const contract = useGetContractQuery(details.applicationId);
	return (
		<Modal title="Условия и документы кредита" size="large" onClose={onClose}>
			<div className={styles.grid}>
				<section>
					<h3>{details.product?.name ?? loan.product.name}</h3>
					<KeyValueList items={[
						{ label: 'Договор', value: details.contractNumber },
						{ label: 'Статус', value: loan.status === 'CLOSED' ? 'Закрыт' : 'Действует' },
						{ label: 'Дата выдачи', value: formatDate(details.issuedAt ?? loan.issuedAt) },
						{ label: 'Сумма', value: formatMoney(details.terms.amount) },
						{ label: 'Остаток основного долга', value: formatMoney(loan.balance) },
						{ label: 'Срок', value: formatTerm(details.terms.term) },
						{ label: 'Ставка', value: formatRate(details.terms.rate) },
						{ label: details.terms.graceMonths ? 'Платёж после отсрочки' : 'Платёж по договору', value: formatMoney(details.terms.monthlyPayment) },
						{ label: 'Отсрочка основного долга', value: details.terms.graceMonths ? formatTerm(details.terms.graceMonths) : 'Не предусмотрена' },
					]} />
					<p>Точные суммы и даты каждого платежа указаны в графике. Последний платёж может отличаться.</p>
				</section>
				<section>
					<h3>Документы договора</h3>
					<AsyncBoundary isLoading={contract.isLoading} isError={contract.isError}
						onRetry={() => void contract.refetch()} loadingLabel="Загрузка документов"
						errorText="Не удалось загрузить документы договора."
						isEmpty={contract.data?.documents.length === 0} emptyTitle="Документов пока нет">
						<ul className={styles.documents}>
							{contract.data?.documents.map((document) => <li key={document.id}>
								<a href={document.url} target="_blank" rel="noreferrer"><Icon name="documents" size={20} />{document.title}<Icon name="arrow" size={16} /></a>
							</li>)}
						</ul>
						<p>Печатные формы банка открываются в новой вкладке. Их можно распечатать или сохранить средствами браузера.</p>
					</AsyncBoundary>
				</section>
			</div>
		</Modal>
	);
};
