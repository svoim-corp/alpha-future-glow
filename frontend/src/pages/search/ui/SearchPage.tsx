import { SEARCH_KINDS, searchKindSchema } from '@entities/search';
import { SearchForm, searchPageTarget, SearchResults } from '@features/global-search';
import { Page, Tabs } from '@shared/ui';
import { useNavigate, useSearchParams } from 'react-router';

export const SearchPage = () => {
	const [params] = useSearchParams();
	const navigate = useNavigate();
	const query = params.get('q') ?? '';
	const parsedType = searchKindSchema.safeParse(params.get('type'));
	const type = parsedType.success ? parsedType.data : undefined;
	return <Page title="Поиск по кабинету" subtitle="Заявки, кредиты, планы, продукты и ответы на вопросы">
		<SearchForm key={query} initialValue={query}
			onSearch={(value) => void navigate(searchPageTarget(value, type))} />
		<Tabs label="Категории поиска" value={type ?? 'all'} items={[
			{ id: 'all', label: 'Все результаты' }, ...SEARCH_KINDS.map((kind) => ({ id: kind.type, label: kind.label })),
		]} onChange={(value) => {
			const next = searchKindSchema.safeParse(value);
			void navigate(searchPageTarget(query, next.success ? next.data : undefined));
		}} />
		<SearchResults key={`${query}:${type ?? 'all'}`} query={query} type={type} />
	</Page>;
};
