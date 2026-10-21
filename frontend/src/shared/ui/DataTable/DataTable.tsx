import clsx from 'clsx';
import { type ReactNode } from 'react';

import styles from './DataTable.module.scss';

export interface DataTableColumn<T> {
	key: string;
	header: string;
	render: (row: T) => ReactNode;
	/** Суммы выравниваем вправо */
	align?: 'left' | 'right';
}

export interface DataTableProps<T> {
	caption: string;
	columns: DataTableColumn<T>[];
	rows: T[];
	getRowKey: (row: T) => string;
	/** Подпись нужна скринридеру, но на экране не показывается (заголовок уже есть рядом) */
	hideCaption?: boolean;
	className?: string;
}

/** Таблица со своим горизонтальным скроллом: на узком экране прокручивается контейнер, а не страница */
export const DataTable = <T,>({ caption, columns, rows, getRowKey, hideCaption, className }: DataTableProps<T>) => (
	<div className={clsx(styles.wrap, className)}>
		<table className={styles.table}>
			<caption className={clsx(hideCaption && 'visually-hidden')}>{caption}</caption>
			<thead>
				<tr>
					{columns.map((column) => (
						<th key={column.key} scope="col" className={clsx(column.align === 'right' && styles.right)}>
							{column.header}
						</th>
					))}
				</tr>
			</thead>
			<tbody>
				{rows.map((row) => (
					<tr key={getRowKey(row)}>
						{columns.map((column) => (
							<td key={column.key} className={clsx(column.align === 'right' && styles.right)}>
								{column.render(row)}
							</td>
						))}
					</tr>
				))}
			</tbody>
		</table>
	</div>
);
