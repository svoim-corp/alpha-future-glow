import { useGetMeQuery } from '@entities/session';
import { Button, Icon, Page } from '@shared/ui';
import { ProductCatalog } from '@widgets/product-catalog';

export const ProductsPage = () => {
	const { data: user } = useGetMeQuery();
	const company = user?.company?.name.match(/«(.+)»/)?.[1] ?? user?.company?.name;

	return (
		<Page
			eyebrow={`КРЕДИТНЫЙ КАБИНЕТ${company ? ` · ${company.toUpperCase()}` : ''}`}
			title="Каталог продуктов"
			subtitle="Выбирайте финансирование под задачу вашего бизнеса."
			actions={(
				<Button
					variant="secondary"
					onClick={() => document.getElementById('calculator')?.scrollIntoView({ behavior: 'smooth' })}
				>
					Рассчитать платёж <Icon name="arrow" size={17} />
				</Button>
			)}
		>
			<ProductCatalog />
		</Page>
	);
};
