import { dashboardConfigSchema } from '../model/schema';
import { type DashboardScope } from '../model/types';

/** Только чтение прежней настройки для явного импорта. Серверная запись всегда приоритетна. */
export const readLegacyDashboardConfig = ({ userId, companyId }: DashboardScope) => {
	try {
		const key = `credit-dashboard:v1:${encodeURIComponent(userId)}:${encodeURIComponent(companyId)}`;
		const raw = localStorage.getItem(key);
		if (raw === null) return null;
		const document: unknown = JSON.parse(raw);
		if (typeof document !== 'object' || document === null || !('config' in document)) throw new Error();
		return dashboardConfigSchema.parse(document.config);
	}
	catch { throw new Error('Не удалось прочитать прежнюю настройку устройства. Настройте главную заново или повторите перенос позже.'); }
};
