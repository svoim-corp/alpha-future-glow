const STATE_LABEL: Record<string, string> = {
	ACTIVE: 'действует',
	LIQUIDATING: 'в процессе ликвидации',
	LIQUIDATED: 'ликвидирована',
	BANKRUPT: 'признана банкротом',
};

/** Состояние компании по реестру словами для клиента; неизвестный код показываем как есть */
export const companyStateLabel = (state: string): string => STATE_LABEL[state] ?? state;

export const isCompanyActive = (state: string): boolean => state === 'ACTIVE';
