import { type SearchItem, type SearchKind } from '@entities/search';
import { paths, ROUTES } from '@shared/config';

export const searchTarget = (type: SearchKind, item: SearchItem): string => {
	const id = encodeURIComponent(item.id);
	switch (type) {
		case 'applications': return item.state === 'DRAFT' ? paths.applicationEdit(id) : paths.application(id);
		case 'loans': return `${ROUTES.credits}?loan=${id}&tab=${item.state === 'CLOSED' ? 'closed' : 'active'}`;
		case 'plans': return `${ROUTES.planner}?plan=${id}`;
		case 'products': return `${ROUTES.products}?product=${id}`;
		case 'faq': return `${ROUTES.support}?faq=${id}`;
		case 'tickets': return `${ROUTES.support}?ticket=${id}`;
	}
};

export const searchPageTarget = (query: string, type?: SearchKind) => {
	const params = new URLSearchParams({ q: query });
	if (type) params.set('type', type);
	return `${ROUTES.search}?${params.toString()}`;
};
