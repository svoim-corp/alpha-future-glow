import { ProductCard, useGetGoalsQuery, useGetProductsQuery, useMatchProductsQuery } from '@entities/product';
import { PaymentEstimate } from '@features/calculate-payment';
import { skipToken } from '@reduxjs/toolkit/query';
import { useDebouncedValue } from '@shared/lib/debounce';
import { AsyncBoundary } from '@shared/ui';
import { useState } from 'react';

import { estimateParams } from '../model/estimate';
import { parseWholeNumber, validateCatalogInput } from '../model/input';
import { ConditionsPanel } from './ConditionsPanel';
import styles from './ProductCatalog.module.scss';
import { TaskPicker } from './TaskPicker';

const DEFAULT_AMOUNT = '1500000';
const DEFAULT_TERM = '24';

export const ProductCatalog = () => {
	const goalsQuery = useGetGoalsQuery();
	const productsQuery = useGetProductsQuery();
	const [pickedGoal, setPickedGoal] = useState<string | null>(null);
	const [amountText, setAmountText] = useState(DEFAULT_AMOUNT);
	const [termText, setTermText] = useState(DEFAULT_TERM);
	const [selected, setSelected] = useState<string | null>(null);

	const goal = pickedGoal ?? goalsQuery.data?.[0]?.code ?? null;
	const errors = validateCatalogInput(amountText, termText);
	const amount = parseWholeNumber(amountText);
	const term = parseWholeNumber(termText);
	const debounced = useDebouncedValue({
		goal,
		amount: errors.amount ? null : amount,
		term: errors.term ? null : term,
	});

	const matchArgs = debounced.goal && debounced.amount !== null && debounced.term !== null
		? { goal: debounced.goal, amount: debounced.amount, term: debounced.term }
		: skipToken;
	const match = useMatchProductsQuery(matchArgs);
	const fitting = new Set((match.data ?? []).filter((item) => item.fits).map((item) => item.code));
	const products = productsQuery.data ?? [];
	const fitNames = products.filter((product) => fitting.has(product.code)).map((product) => product.name);
	const fitLine = match.data
		? (fitNames.length > 0
			? `Подходят: ${fitNames.join(', ')}. Сумма и срок перейдут в форму заявки.`
			: 'Под эту задачу, сумму и срок продуктов не нашлось. Измените сумму или срок.')
		: null;

	return (
		<div className={styles.layout}>
			<div className={styles.main}>
				<AsyncBoundary
					isLoading={goalsQuery.isLoading}
					isError={goalsQuery.isError}
					onRetry={() => void goalsQuery.refetch()}
					loadingLabel="Загрузка целей"
					errorText="Не удалось загрузить список задач."
				>
					<TaskPicker
						goals={goalsQuery.data ?? []}
						goal={goal}
						amountText={amountText}
						termText={termText}
						errors={errors}
						fitLine={fitLine}
						onGoalChange={setPickedGoal}
						onAmountChange={setAmountText}
						onTermChange={setTermText}
					/>
				</AsyncBoundary>
				<section aria-labelledby="products-title">
					<div className={styles.listHead}>
						<h2 id="products-title" className={styles.sectionTitle}>Все продукты</h2>
						<p className={styles.sub}>Платёж показан для суммы и срока из подбора</p>
					</div>
					<AsyncBoundary
						isLoading={productsQuery.isLoading}
						isError={productsQuery.isError}
						isEmpty={products.length === 0}
						onRetry={() => void productsQuery.refetch()}
						loadingLabel="Загрузка продуктов"
						errorText="Не удалось загрузить продукты."
						emptyTitle="Продуктов пока нет"
					>
						<ul className={styles.grid} aria-label="Продукты">
							{products.map((product) => (
								<li key={product.code}>
									<ProductCard
										product={product}
										selected={selected === product.code}
										fits={fitting.has(product.code)}
										onSelect={setSelected}
										estimate={product.term ? (
											<PaymentEstimate
												params={estimateParams(product.code, amount, term)}
												compact
											/>
										) : (
											<p className={styles.sub}>Проценты только за дни в минусе</p>
										)}
									/>
								</li>
							))}
						</ul>
					</AsyncBoundary>
				</section>
			</div>
			<ConditionsPanel productCode={selected} amount={amount} term={term} goal={goal} />
		</div>
	);
};
