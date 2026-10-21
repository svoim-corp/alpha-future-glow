import { type ChangeEvent, type ReactNode, useId } from 'react';

import { Icon } from '../Icon';
import { Select, type SelectOption } from '../Select';
import styles from './Toolbar.module.scss';

export interface SearchFieldProps {
	value: string;
	onChange: (value: string) => void;
	label: string;
	placeholder?: string;
}

/** Локальный поиск в панели фильтров списка */
export const SearchField = ({ value, onChange, label, placeholder }: SearchFieldProps) => (
	<label className={styles.search}>
		<Icon name="search" size={18} />
		<input
			type="search"
			aria-label={label}
			placeholder={placeholder}
			value={value}
			onChange={(event: ChangeEvent<HTMLInputElement>) => onChange(event.target.value)}
		/>
	</label>
);

export interface SelectFilterProps {
	label: string;
	options: SelectOption[];
	value: string;
	onChange: (value: string) => void;
}

/** Компактный фильтр-список с подписью над значением */
export const SelectFilter = ({ label, options, value, onChange }: SelectFilterProps) => {
	const labelId = useId();

	return (
		<div className={styles.filter}>
			<span id={labelId}>{label}</span>
			<Select variant="filter" aria-labelledby={labelId} options={options} value={value} onChange={onChange} />
		</div>
	);
};

/** Панель фильтров списка: поиск и выпадающие фильтры в одну строку */
export const FilterToolbar = ({ children }: { children: ReactNode }) => (
	<div className={styles.toolbar}>{children}</div>
);

/** Строка под фильтрами: счётчик результатов слева, сортировка справа */
export const ListToolbar = ({ children }: { children: ReactNode }) => (
	<div className={styles.list}>{children}</div>
);
