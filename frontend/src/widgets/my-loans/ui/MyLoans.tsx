import { LoanCard, PaymentsTable, useGetLoanQuery, useGetLoansQuery } from '@entities/loan';
import { ROUTES } from '@shared/config';
import { formatMoney, formatRate, formatTerm } from '@shared/lib/format';
import { AsyncBoundary, Button, ButtonLink, KeyValueList, Modal } from '@shared/ui';
import { useState } from 'react';

import styles from './MyLoans.module.scss';

const LoanDetailsModal = ({ loanId, onClose }: { loanId: string; onClose: () => void }) => {
	const { data, isLoading, isError, refetch } = useGetLoanQuery(loanId);

	return (
		<Modal title="Детали кредита" onClose={onClose} wide>
			<AsyncBoundary
				isLoading={isLoading}
				isError={isError}
				onRetry={() => void refetch()}
				loadingLabel="Загружаем договор"
				errorText="Не удалось загрузить детали кредита"
			>
				{data && (
					<>
						<KeyValueList
							items={[
								{ label: 'Договор', value: data.contractNumber },
								{ label: 'Сумма', value: formatMoney(data.terms.amount) },
								{ label: 'Срок', value: formatTerm(data.terms.term) },
								{ label: 'Ставка', value: formatRate(data.terms.rate) },
								{ label: 'Ежемесячный платёж', value: formatMoney(data.terms.monthlyPayment) },
							]}
						/>
						<PaymentsTable payments={data.payments} />
					</>
				)}
			</AsyncBoundary>
		</Modal>
	);
};

export const MyLoans = () => {
	const { data, isLoading, isError, refetch } = useGetLoansQuery();
	const [openId, setOpenId] = useState<string | null>(null);

	return (
		<AsyncBoundary
			isLoading={isLoading}
			isError={isError}
			isEmpty={data?.length === 0}
			onRetry={() => void refetch()}
			loadingLabel="Загружаем кредиты"
			errorText="Не удалось загрузить кредиты"
			emptyTitle="Выданных кредитов пока нет"
			emptyText="Когда банк выдаст кредит, он появится здесь вместе с графиком платежей."
			emptyAction={<ButtonLink to={ROUTES.products}>Подобрать финансирование</ButtonLink>}
		>
			<div className={styles.list}>
				{data?.map((loan) => (
					<LoanCard
						key={loan.id}
						loan={loan}
						actions={(
							<div>
								<Button variant="secondary" onClick={() => setOpenId(loan.id)}>
									{`Условия и график: ${loan.product.name}`}
								</Button>
							</div>
						)}
					/>
				))}
			</div>
			{openId && <LoanDetailsModal loanId={openId} onClose={() => setOpenId(null)} />}
		</AsyncBoundary>
	);
};
