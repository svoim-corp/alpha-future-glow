import { normalizeSearch, type SearchKind, useGlobalSearchQuery } from '@entities/search';
import { useGetMeQuery } from '@entities/session';
import { skipToken } from '@reduxjs/toolkit/query';

export const useSearchData = (query: string, type?: SearchKind, cursor?: string, limit = 3, enabled = true) => {
	const profile = useGetMeQuery();
	const user = profile.data;
	const normalized = normalizeSearch(query);
	const length = Array.from(normalized).length;
	const valid = length >= 2 && length <= 120;
	const request = useGlobalSearchQuery(enabled && valid && user?.role === 'CLIENT' && user.company ? {
		query: normalized, type, cursor, limit, userId: user.id, companyId: user.company.id,
	} : skipToken, { refetchOnMountOrArgChange: true, refetchOnFocus: true });
	const accessError = profile.isError ? 'Не удалось загрузить профиль для поиска.'
		: user && (user.role !== 'CLIENT' || !user.company) ? 'Поиск доступен клиенту с подключённой компанией.' : null;
	return { ...request, valid, normalized, accessError, profileLoading: profile.isLoading };
};
