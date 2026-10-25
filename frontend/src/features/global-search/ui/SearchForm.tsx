import { normalizeSearch } from '@entities/search';
import { Button, SearchField } from '@shared/ui';
import { useState } from 'react';

import styles from './Search.module.scss';

interface SearchFormProps { initialValue: string; onSearch: (query: string) => void }

export const SearchForm = ({ initialValue, onSearch }: SearchFormProps) => {
	const [query, setQuery] = useState(initialValue);
	return <form className={styles.form} role="search" onSubmit={(event) => {
		event.preventDefault(); onSearch(normalizeSearch(query));
	}}>
		<SearchField label="Что найти в кабинете" value={query} onChange={setQuery} placeholder="Название или номер" />
		<Button type="submit">Найти</Button>
	</form>;
};
