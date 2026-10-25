import { normalizeSearch, SEARCH_KINDS } from '@entities/search';
import { getErrorMessage } from '@shared/api';
import { useDebouncedValue } from '@shared/lib/debounce';
import { Button, Icon } from '@shared/ui';
import { Fragment, useEffect, useId, useRef, useState } from 'react';
import { Link, useLocation, useNavigate } from 'react-router';

import { searchStateLabel } from '../model/stateLabel';
import { searchPageTarget, searchTarget } from '../model/target';
import { useSearchData } from '../model/useSearchData';
import styles from './Search.module.scss';

interface SearchSection { label: string; to: string; keywords?: readonly string[] }
interface GlobalSearchProps { sections: readonly SearchSection[]; onOpen?: () => void }

const SearchInput = ({ sections, onOpen }: GlobalSearchProps) => {
	const [query, setQuery] = useState('');
	const [open, setOpen] = useState(false);
	const [activeKey, setActiveKey] = useState<string | null>(null);
	const root = useRef<HTMLDivElement>(null);
	const input = useRef<HTMLInputElement>(null);
	const listId = useId();
	const navigate = useNavigate();
	const normalized = normalizeSearch(query);
	const valid = Array.from(normalized).length >= 2 && Array.from(normalized).length <= 120;
	const debounced = useDebouncedValue(normalized, 300);
	const result = useSearchData(debounced, undefined, undefined, 3, open && valid && normalized === debounced);
	const waiting = normalized !== debounced;
	const groups = !waiting && !result.isError && result.currentData?.query === normalized
		? result.currentData.groups : [];
	const navigation = sections.filter((section) => [section.label, ...(section.keywords ?? [])]
		.some((label) => normalizeSearch(label).includes(normalized)));
	const options = [
		...navigation.map((section) => ({ key: `section:${section.to}`, title: section.label, snippet: 'Раздел кабинета', to: section.to })),
		...groups.flatMap((group) => group.items.map((item) => ({ key: `${group.type}:${item.id}`,
			title: item.title, snippet: [searchStateLabel(group.type, item.state), item.snippet].filter(Boolean).join(' · '), to: searchTarget(group.type, item) }))),
	];
	const active = options.find((option) => option.key === activeKey);
	const choose = (to: string) => {
		setOpen(false); setQuery(''); setActiveKey(null);
		void navigate(to);
	};
	useEffect(() => {
		if (!open) return undefined;
		const outside = (event: PointerEvent) => {
			if (event.target instanceof Node && !root.current?.contains(event.target)) setOpen(false);
		};
		document.addEventListener('pointerdown', outside);
		return () => document.removeEventListener('pointerdown', outside);
	}, [open]);
	useEffect(() => {
		const shortcut = (event: KeyboardEvent) => {
			if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 'k') {
				event.preventDefault(); input.current?.focus();
			}
		};
		document.addEventListener('keydown', shortcut);
		return () => document.removeEventListener('keydown', shortcut);
	}, []);
	const visible = open;
	const renderOption = (option: typeof options[number]) => <button type="button" role="option" tabIndex={-1}
		id={`${listId}-${encodeURIComponent(option.key)}`} key={option.key} className={styles.option}
		aria-selected={activeKey === option.key} onPointerMove={() => setActiveKey(option.key)}
		onClick={() => choose(option.to)}><strong>{option.title}</strong><small>{option.snippet}</small></button>;
	return <div ref={root} className={styles.root}>
		<label className={styles.input}>
			<Icon name="search" size={20} /><span className="visually-hidden">Поиск по кабинету</span>
			<input ref={input} type="search" role="combobox" autoComplete="off" placeholder="Заявки, кредиты, планы, вопросы…"
				value={query} aria-expanded={visible} aria-controls={visible ? listId : undefined} aria-autocomplete="list"
				aria-activedescendant={visible && active ? `${listId}-${encodeURIComponent(active.key)}` : undefined}
				onFocus={() => { setOpen(true); onOpen?.(); }}
				onChange={(event) => { setQuery(event.target.value); setOpen(true); setActiveKey(null); }}
				onKeyDown={(event) => {
					if (event.nativeEvent.isComposing) return;
					if (event.key === 'Escape') { setOpen(false); setActiveKey(null); }
					if (event.key === 'Tab') setOpen(false);
					if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
						event.preventDefault(); setOpen(true);
						if (options.length === 0) return;
						const index = options.findIndex((option) => option.key === activeKey);
						const next = event.key === 'ArrowDown'
							? (index + 1) % options.length
							: (index <= 0 ? options.length - 1 : index - 1);
						const option = options[next];
						setActiveKey(option?.key ?? null);
						if (option) document.getElementById(`${listId}-${encodeURIComponent(option.key)}`)
							?.scrollIntoView({ block: 'nearest' });
					}
					if (event.key === 'Enter') {
						event.preventDefault();
						if (active) choose(active.to);
						else if (valid) choose(searchPageTarget(normalized));
					}
				}} />
		</label>
		{visible && <div className={styles.panel}>
			<div id={listId} role="listbox" aria-label="Результаты поиска">
				{navigation.length > 0 && <div role="group" aria-label="Разделы кабинета">
					<p className={styles.caption}>Разделы</p>{options.slice(0, navigation.length).map(renderOption)}
				</div>}
				{groups.map((group) => <div role="group" key={group.type}
					aria-label={SEARCH_KINDS.find((kind) => kind.type === group.type)?.label}>
					{group.items.length > 0 && <p className={styles.caption}>
						{SEARCH_KINDS.find((kind) => kind.type === group.type)?.label}
					</p>}
					{group.items.map((item) => {
						const option = options.find((entry) => entry.key === `${group.type}:${item.id}`);
						return option ? <Fragment key={option.key}>{renderOption(option)}</Fragment> : null;
					})}
				</div>)}
			</div>
			{!normalized && <p className={styles.hint}>
				Введите название или номер. Поиск по данным — от двух символов.
			</p>}
			{valid && result.accessError && <p className={styles.hint} role="alert">{result.accessError}</p>}
			{normalized && !valid && <p className={styles.hint}>Введите от 2 до 120 символов.</p>}
			{valid && (waiting || result.isFetching) && <p className={styles.hint} role="status">
				Ищем в кабинете…
			</p>}
			{valid && !waiting && result.isError && <div className={styles.hint} role="alert">
				{getErrorMessage(result.error, 'Поиск временно недоступен.')}
				<Button variant="ghost" compact onClick={() => void result.refetch()}>Повторить</Button>
			</div>}
			{groups.some((group) => group.unavailable) && <p className={styles.hint} role="status">
				Временно недоступны: {groups.filter((group) => group.unavailable)
					.map((group) => SEARCH_KINDS.find((kind) => kind.type === group.type)?.label).join(', ')}. Выдача неполная.
			</p>}
			{!waiting && !result.isFetching && !result.isError && valid && groups.length > 0
				&& options.length === 0 && groups.every((group) => !group.unavailable)
				&& <p className={styles.hint}>Ничего не найдено. Попробуйте другое название или номер.</p>}
			{valid && !waiting && <div className={styles.footer}>
				<Link to={searchPageTarget(normalized)} onClick={() => setOpen(false)}>Все результаты поиска →</Link>
			</div>}
		</div>}
	</div>;
};

export const GlobalSearch = (props: GlobalSearchProps) => {
	const location = useLocation();
	return <SearchInput key={location.key} {...props} />;
};
