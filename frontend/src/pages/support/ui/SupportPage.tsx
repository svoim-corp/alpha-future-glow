import { useGetMeQuery } from '@entities/session';
import { Badge, Page } from '@shared/ui';
import { SupportCenter } from '@widgets/support-center';

export const SupportPage = () => {
	const { data: user } = useGetMeQuery();
	const company = user?.company?.name.match(/«(.+)»/)?.[1] ?? user?.company?.name;

	return (
		<Page
			eyebrow={`КРЕДИТНЫЙ КАБИНЕТ${company ? ` · ${company.toUpperCase()}` : ''}`}
			title="Поддержка"
			subtitle="Разберёмся вместе — от первого вопроса до решения."
			actions={<Badge>Помощь внутри кабинета</Badge>}
		>
			<SupportCenter />
		</Page>
	);
};
