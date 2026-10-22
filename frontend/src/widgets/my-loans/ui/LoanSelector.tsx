import { formatMoney } from '@shared/lib/format';
import { Icon } from '@shared/ui';
import { useEffect, useRef } from 'react';

import styles from './LoanSelector.module.scss';

interface LoanSelectorItem {
	id: string;
	title: string;
	caption: string;
	amount?: number;
}

interface LoanSelectorProps {
	items: LoanSelectorItem[];
	value: string;
	onChange: (id: string) => void;
	detailsId: string;
}

export const LoanSelector = ({ items, value, onChange, detailsId }: LoanSelectorProps) => {
	const trackRef = useRef<HTMLDivElement>(null);

	useEffect(() => {
		const track = trackRef.current;
		const selected = track?.querySelector<HTMLButtonElement>('[aria-pressed="true"]');
		if (!track || !selected) return;
		const bounds = track.getBoundingClientRect();
		const card = selected.getBoundingClientRect();
		const offset = card.left < bounds.left ? card.left - bounds.left - 4
			: card.right > bounds.right ? card.right - bounds.right + 4 : 0;
		if (offset) track.scrollBy({ left: offset, behavior: 'instant' });
	}, [value]);

	return (
		<div className={styles.root}>
			<div className={styles.heading}>
				<span>Выберите кредит</span>
				<span className={styles.counter}>{items.findIndex((item) => item.id === value) + 1} из {items.length}</span>
			</div>
			<div ref={trackRef} className={styles.track} role="group" aria-label="Выбор кредита">
				{items.map((item) => (
					<button key={item.id} type="button" className={styles.card}
						aria-pressed={item.id === value} aria-controls={detailsId} onClick={() => onChange(item.id)}>
						<span className={styles.top}>
							<span className={styles.title} title={item.title}>{item.title}</span>
							<span className={styles.marker} aria-hidden="true">
								{item.id === value && <Icon name="check" size={14} />}
							</span>
						</span>
						<span className={styles.caption}>{item.caption}</span>
						<span className={styles.amount}>{item.amount === undefined ? 'Условия не загружены' : formatMoney(item.amount)}</span>
					</button>
				))}
			</div>
		</div>
	);
};
