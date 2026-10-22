import { useGetProductQuery } from '@entities/product';
import { Alert, Button } from '@shared/ui';
import { useId } from 'react';

import styles from './PlannerBoard.module.scss';

interface PlanFundingProps {
	productCode: string;
	productName: string;
}

export const PlanFunding = ({ productCode, productName }: PlanFundingProps) => {
	const product = useGetProductQuery(productCode);
	const hintId = useId();

	return (
		<div className={styles.funding}>
			<h3>{product.currentData?.name ?? productName}</h3>
			{product.isLoading || product.isFetching ? <p role="status">Загружаем описание кредита…</p>
				: product.isError || !product.currentData ? (
					<Alert tone="warning" title="Описание кредита пока недоступно">
						<Button variant="ghost" compact onClick={() => void product.refetch()}>Повторить загрузку</Button>
					</Alert>
				) : <p>{product.currentData.description}</p>}
			<Button disabled aria-describedby={hintId}>Перейти к оформлению</Button>
			<small id={hintId}>Страница оформления пока недоступна.</small>
		</div>
	);
};
