import { type ReactNode } from 'react';

import styles from './KeyValueList.module.scss';

export interface KeyValueItem {
	label: string;
	value: ReactNode;
}

export interface KeyValueListProps {
	items: KeyValueItem[];
}

export const KeyValueList = ({ items }: KeyValueListProps) => (
	<dl className={styles.list}>
		{items.map((item) => (
			<div key={item.label} className={styles.row}>
				<dt>{item.label}</dt>
				<dd>{item.value}</dd>
			</div>
		))}
	</dl>
);
