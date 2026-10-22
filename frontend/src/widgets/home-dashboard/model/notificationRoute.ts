/** Уведомления ведут только к разделам клиентского кабинета, без внешних URL. */
export const notificationRoute = (link: string): string | null => {
	if (!link.startsWith('/') || link.startsWith('//') || /[\\\u0000-\u001f]/u.test(link)) return null;
	try {
		const url = new URL(link, 'https://cabinet.invalid');
		const section = url.pathname.split('/')[1] ?? '';
		if (url.origin !== 'https://cabinet.invalid' || url.pathname.startsWith('//')
			|| !['', 'applications', 'credits', 'planner', 'products', 'support'].includes(section)) return null;
		return `${url.pathname}${url.search}${url.hash}`;
	} catch { return null; }
};
