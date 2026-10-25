import { SEARCH_KINDS, type SearchKind } from '@entities/search';
import { getErrorMessage } from '@shared/api';
import { Alert, Button } from '@shared/ui';
import { useState } from 'react';
import { Link } from 'react-router';

import { searchStateLabel } from '../model/stateLabel';
import { searchPageTarget, searchTarget } from '../model/target';
import { useSearchData } from '../model/useSearchData';
import styles from './Search.module.scss';

interface SearchResultsProps { query: string; type?: SearchKind }

export const SearchResults = ({ query, type }: SearchResultsProps) => {
	const [positions, setPositions] = useState<(string | undefined)[]>([undefined]);
	const result = useSearchData(query, type, positions[positions.length - 1], type ? 20 : 3);
	if (!result.valid) return <p className={styles.hint}>Введите от 2 до 120 символов.</p>;
	if (result.accessError) return <Alert tone="danger" title={result.accessError} />;
	if (result.profileLoading) return <p role="status">Загружаем профиль…</p>;
	if (result.isLoading || (result.isFetching && !result.currentData)) return <p role="status">Ищем в кабинете…</p>;
	if (result.isError) return <Alert tone="danger" title={getErrorMessage(result.error, 'Не удалось выполнить поиск.')}
		actions={<>
			<Button variant="secondary" onClick={() => void result.refetch()}>Повторить</Button>
			{positions.length > 1 && <Button variant="ghost" onClick={() => setPositions([undefined])}>
				К первой странице
			</Button>}
		</>} />;
	const groups = result.currentData?.groups ?? [];
	const visibleGroups = groups.filter((group) => type || group.items.length > 0 || group.unavailable);
	const empty = groups.length > 0 && groups.every((group) => !group.unavailable && group.items.length === 0);
	return <div className={styles.results} aria-busy={result.isFetching}>
		{empty && <p>Ничего не найдено. Попробуйте другое название или номер.</p>}
		{visibleGroups.map((group) => <section key={group.type} className={styles.group}>
			<h2>{SEARCH_KINDS.find((kind) => kind.type === group.type)?.label}</h2>
			{group.unavailable ? <Alert tone="warning" title="Поиск по этому разделу временно недоступен."
				actions={<Button variant="ghost" onClick={() => void result.refetch()}>Повторить</Button>} />
				: group.items.length > 0 ? <ul className={styles.list}>{group.items.map((item) => <li key={item.id}>
					<Link className={styles.option} to={searchTarget(group.type, item)}>
						<strong>{item.title}</strong>
						<small>{item.id} · {searchStateLabel(group.type, item.state)}</small>
						{item.snippet && <small>{item.snippet}</small>}
					</Link>
				</li>)}</ul> : <p className={styles.hint}>Совпадений нет.</p>}
			{!type && group.hasMore && <Link to={searchPageTarget(query, group.type)}>Все результаты в разделе →</Link>}
			{type && <nav className={styles.navigation} aria-label="Страницы результатов поиска">
				<Button variant="secondary" disabled={positions.length === 1 || result.isFetching}
					onClick={() => setPositions((current) => current.slice(0, -1))}>Назад</Button>
				<span aria-live="polite">Страница {positions.length}</span>
				<Button variant="secondary" disabled={!group.hasMore || !group.nextCursor || result.isFetching}
					onClick={() => {
						const next = group.nextCursor;
						if (next) setPositions((current) => [...current, next]);
					}}>Далее</Button>
			</nav>}
		</section>)}
	</div>;
};
