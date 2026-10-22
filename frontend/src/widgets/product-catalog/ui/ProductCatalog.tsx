import {
	ProductCard,
	useGetGoalsQuery,
	useGetProductsQuery,
	useMatchProductsQuery,
} from '@entities/product';
import { skipToken } from '@reduxjs/toolkit/query';
import { ROUTES } from '@shared/config';
import { useDebouncedValue } from '@shared/lib/debounce';
import { Alert, AsyncBoundary, Button, Icon } from '@shared/ui';
import clsx from 'clsx';
import { useState } from 'react';
import { Link } from 'react-router';

import { clamp, DEFAULT_AMOUNT, DEFAULT_TERM } from '../model/range';
import { selectionAdjustment,selectProduct } from '../model/selection';
import { ConditionsDialog } from './ConditionsDialog';
import { PaymentCalculator } from './PaymentCalculator';
import styles from './ProductCatalog.module.scss';

/** Каталог по макету catalog.html: баннер, цели-фильтры, карточки одной высоты, калькулятор, подсказка */
export const ProductCatalog = () => {
	const goalsQuery = useGetGoalsQuery();
	const productsQuery = useGetProductsQuery();
	const [goal, setGoal] = useState<string | null>(null);
	const [calcCode, setCalcCode] = useState<string | null>(null);
	const [amountValue, setAmountValue] = useState(DEFAULT_AMOUNT);
	const [termValue, setTermValue] = useState(DEFAULT_TERM);
	const [opened, setOpened] = useState<string | null>(null);

	const products = productsQuery.data ?? [];
	const eligible = goal ? products.filter((product) => product.goals.includes(goal)) : products;
	const calcProduct = selectProduct(products, calcCode, goal);
	const amount = calcProduct ? clamp(amountValue, calcProduct.amount) : amountValue;
	const term = calcProduct?.term ? clamp(termValue, calcProduct.term) : null;
	const debouncedGoal = useDebouncedValue(goal);
	const debouncedAmount = useDebouncedValue(amount);
	const debouncedTerm = useDebouncedValue(term ?? 0);
	const debounced = { goal: debouncedGoal, amount: debouncedAmount, term: debouncedTerm };
	const match = useMatchProductsQuery(
		debounced.goal
			? { goal: debounced.goal, amount: debounced.amount, term: debounced.term }
			: skipToken,
	);
	const waiting =
		goal !== debouncedGoal
		|| amount !== debouncedAmount
		|| (term ?? 0) !== debouncedTerm
		|| match.isFetching;
	const results = goal && !waiting && !match.isError ? match.currentData : undefined;
	const fitting = new Set((results ?? []).filter((item) => item.fits).map((item) => item.code));
	const openedProduct = products.find((item) => item.code === opened);
	const order = new Map((results ?? []).map((item, index) => [item.code, index]));
	const shown = [...eligible].sort(
		(a, b) => (order.get(a.code) ?? products.length) - (order.get(b.code) ?? products.length),
	);
	const fitNames = shown
		.filter((product) => fitting.has(product.code))
		.map((product) => product.name);
	const fitLine = results
		? fitNames.length > 0
			? `Подходят: ${fitNames.join(', ')}. Сумма и срок перейдут в форму заявки.`
			: 'Под эту задачу, сумму и срок продуктов не нашлось. Измените сумму или срок в калькуляторе.'
		: null;

	return (
		<div className={styles.root}>
			<section className={styles.banner} aria-labelledby="catalog-banner-title">
				<div>
					<span className={styles.bannerLabel}>ОТ ЗАДАЧИ — К РЕШЕНИЮ</span>
					<h2 id="catalog-banner-title">
						У бизнеса есть планы.
						<br />У нас — варианты.
					</h2>
					<p>Начните с цели, сравните условия и оцените платёж.</p>
				</div>
				<div className={styles.art} aria-hidden="true">
					<span />
					<span />
					<span />
				</div>
			</section>
			<AsyncBoundary
				isLoading={goalsQuery.isLoading}
				isError={goalsQuery.isError}
				onRetry={() => void goalsQuery.refetch()}
				loadingLabel="Загрузка целей"
				errorText="Не удалось загрузить список задач."
			>
				<div className={styles.chips} role="group" aria-label="Цель финансирования">
					{[{ code: null, name: 'Все предложения' }, ...(goalsQuery.data ?? [])].map(
						(item) => (
							<button
								key={item.code ?? 'all'}
								type="button"
								aria-pressed={goal === item.code}
								className={clsx(
									styles.chip,
									goal === item.code && styles.chipActive,
								)}
								onClick={() => setGoal(item.code)}
							>
								{item.name}
							</button>
						),
					)}
				</div>
			</AsyncBoundary>
			<AsyncBoundary
				isLoading={productsQuery.isLoading}
				isError={productsQuery.isError}
				isEmpty={products.length === 0}
				onRetry={() => void productsQuery.refetch()}
				loadingLabel="Загрузка продуктов"
				errorText="Не удалось загрузить продукты."
				emptyTitle="Продуктов пока нет"
			>
				{goal && waiting && (
					<p className={styles.fitLine} role="status">
						Подбираем продукты по цели, сумме и сроку…
					</p>
				)}
				{goal && !waiting && match.isError && (
					<Alert
						tone="danger"
						title="Не удалось подобрать продукты."
						actions={
							<Button
								variant="secondary"
								compact
								onClick={() => void match.refetch()}
							>
								Повторить подбор
							</Button>
						}
					>
						Условия продуктов доступны ниже. Соответствие выбранным параметрам пока не
						подтверждено.
					</Alert>
				)}
				{fitLine && (
					<p className={styles.fitLine} role="status">
						{fitLine}
					</p>
				)}
				{shown.length === 0 && (
					<Alert title="Для этой цели продуктов пока нет.">
						Выберите другую цель или «Все предложения».
					</Alert>
				)}
				<ul className={styles.grid} aria-label="Продукты">
					{shown.map((product) => (
						<li key={product.code}>
							<ProductCard
								product={product}
								fits={fitting.has(product.code)}
								reason={results?.find((item) => item.code === product.code)?.reason}
								goalLabel={
									goalsQuery.data?.find(
										(item) => item.code === (goal ?? product.goals[0]),
									)?.name
								}
								onSelect={setOpened}
							/>
						</li>
					))}
				</ul>
				{calcProduct && (
					<PaymentCalculator
						products={eligible}
						product={calcProduct}
						amount={amount}
						term={term}
						goal={goal}
						adjustment={selectionAdjustment(amountValue, termValue, calcProduct)}
						onProductChange={setCalcCode}
						onAmountChange={setAmountValue}
						onTermChange={setTermValue}
					/>
				)}
			</AsyncBoundary>
			<section className={styles.help} aria-label="Помощь с выбором">
				<span className={styles.helpIcon}>
					<Icon name="support" size={20} />
				</span>
				<div>
					<h2>Не уверены, что подойдёт?</h2>
					<p>Расскажите о задаче — поможем разобраться в вариантах.</p>
				</div>
				<Link to={ROUTES.support} className={styles.helpLink}>
					Задать вопрос <Icon name="arrow" size={14} />
				</Link>
			</section>
			{openedProduct && (
				<ConditionsDialog
					key={openedProduct.code}
					productCode={openedProduct.code}
					amount={amount}
					term={term}
					goal={goal}
					onClose={() => setOpened(null)}
				/>
			)}
		</div>
	);
};
