import { STATUS_VIEW } from '@entities/application';
import { PLAN_STATE_LABELS } from '@entities/plan';
import { type SearchKind } from '@entities/search';

export const searchStateLabel = (type: SearchKind, state: string): string => {
	if (type === 'applications') return Object.entries(STATUS_VIEW).find(([code]) => code === state)?.[1].label ?? '';
	if (type === 'plans') return Object.entries(PLAN_STATE_LABELS).find(([code]) => code === state)?.[1] ?? '';
	if (type === 'loans') return state === 'CLOSED' ? 'Закрыт' : state === 'ACTIVE' ? 'Действует' : '';
	if (type === 'tickets') {
		return [{ code: 'OPEN', label: 'Открыто' }, { code: 'ANSWERED', label: 'Есть ответ' },
			{ code: 'CLOSED', label: 'Закрыто' }].find((entry) => entry.code === state)?.label ?? '';
	}
	return '';
};
