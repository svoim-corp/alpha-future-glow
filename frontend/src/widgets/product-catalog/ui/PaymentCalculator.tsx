import { type ProductSummary } from '@entities/product';
import { PaymentEstimate } from '@features/calculate-payment';
import { StartApplicationButton } from '@features/start-application';
import { formatCompactMoney, formatMoney, formatTerm } from '@shared/lib/format';
import { Icon } from '@shared/ui';

import { estimateParams } from '../model/estimate';
import styles from './PaymentCalculator.module.scss';
import { RangeField } from './RangeField';

interface PaymentCalculatorProps {
	products: ProductSummary[];
	product: ProductSummary;
	amount: number;
	/** null у продуктов без срока (овердрафт) */
	term: number | null;
	goal: string | null;
	onProductChange: (code: string) => void;
	onAmountChange: (amount: number) => void;
	onTermChange: (term: number) => void;
}

const AMOUNT_STEP = 50000;

/** «Платёж под ваши планы» по макету catalog.html: продукт, сумма и срок → примерный платёж с сервера */
export const PaymentCalculator = ({
	products,
	product,
	amount,
	term,
	goal,
	onProductChange,
	onAmountChange,
	onTermChange,
}: PaymentCalculatorProps) => (
	<section className={styles.calculator} id="calculator" aria-labelledby="calculator-title">
		<div className={styles.heading}>
			<div>
				<span className={styles.eyebrow}>ПОСЧИТАЕМ ПРЕДВАРИТЕЛЬНО</span>
				<h2 id="calculator-title">Платёж под ваши планы</h2>
				<p>Пример расчёта для продукта «{product.name}»</p>
			</div>
			<Icon name="catalog" size={22} />
		</div>
		<div className={styles.grid}>
			<div className={styles.controls}>
				<RangeField
					label={product.term ? 'Сумма кредита' : 'Лимит'}
					display={formatMoney(amount)}
					value={amount}
					range={product.amount}
					step={AMOUNT_STEP}
					minLabel={formatCompactMoney(product.amount.min)}
					maxLabel={formatCompactMoney(product.amount.max)}
					onChange={onAmountChange}
				/>
				{product.term && term !== null && (
					<RangeField
						label="Срок"
						display={formatTerm(term)}
						value={term}
						range={product.term}
						step={1}
						minLabel={formatTerm(product.term.min)}
						maxLabel={formatTerm(product.term.max)}
						onChange={onTermChange}
					/>
				)}
				<label className={styles.select}>
					Продукт
					<select value={product.code} onChange={(event) => onProductChange(event.target.value)}>
						{products.map((item) => <option key={item.code} value={item.code}>{item.name}</option>)}
					</select>
				</label>
			</div>
			<div className={styles.result}>
				<span>Примерный ежемесячный платёж</span>
				<div className={styles.payment} aria-live="polite">
					{product.term ? (
						<PaymentEstimate params={estimateParams(product.code, amount, term)} />
					) : (
						<p className={styles.hint}>Проценты только за дни в минусе</p>
					)}
				</div>
				<StartApplicationButton
					productCode={product.code}
					amount={amount}
					term={product.term ? term : null}
					goal={goal ?? undefined}
				/>
			</div>
		</div>
		<p className={styles.note}>
			Расчёт по ставке продукта, без комиссий и дополнительных услуг. Это не окончательные условия и не полная
			стоимость кредита. Итог определится после рассмотрения заявки.
		</p>
	</section>
);
