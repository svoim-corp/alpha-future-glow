import clsx from 'clsx';
import { type KeyboardEvent, useRef } from 'react';

import styles from './Tabs.module.scss';

export interface TabItem<T extends string> {
	id: T;
	label: string;
	count?: number;
}

export interface TabsProps<T extends string> {
	items: TabItem<T>[];
	value: T;
	onChange: (id: T) => void;
	label: string;
}

/** id вкладки и панели связаны: вкладка `tab-<id>`, панель `panel-<id>` (панель рисует вызывающий код) */
export const tabId = (id: string) => `tab-${id}`;
export const panelId = (id: string) => `panel-${id}`;

/** Вкладки страницы: переключение стрелками, подсчёт в бейдже (макеты 03–07) */
export const Tabs = <T extends string>({ items, value, onChange, label }: TabsProps<T>) => {
	const refs = useRef(new Map<T, HTMLButtonElement>());

	const onKeyDown = (event: KeyboardEvent<HTMLButtonElement>, index: number) => {
		const step = event.key === 'ArrowRight' ? 1 : event.key === 'ArrowLeft' ? -1 : 0;

		if (step === 0) {
			return;
		}

		event.preventDefault();
		const next = items[(index + step + items.length) % items.length];

		if (next) {
			onChange(next.id);
			refs.current.get(next.id)?.focus();
		}
	};

	return (
		<div className={styles.tabs} role="tablist" aria-label={label}>
			{items.map((item, index) => (
				<button
					key={item.id}
					type="button"
					role="tab"
					id={tabId(item.id)}
					aria-selected={item.id === value}
					aria-controls={panelId(item.id)}
					tabIndex={item.id === value ? 0 : -1}
					className={clsx(styles.tab)}
					ref={(node) => {
						if (node) {
							refs.current.set(item.id, node);
						}
					}}
					onClick={() => onChange(item.id)}
					onKeyDown={(event) => onKeyDown(event, index)}
				>
					{item.label}
					{item.count !== undefined && <span>{item.count}</span>}
				</button>
			))}
		</div>
	);
};
