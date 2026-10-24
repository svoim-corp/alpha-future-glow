import { Page } from '@shared/ui';
import { ProductCatalog } from '@widgets/product-catalog';

export const ProductsPage = () => {
	return (
		<Page
			title="Каталог продуктов"
			subtitle="Выбирайте финансирование под задачу вашего бизнеса."
		>
			<ProductCatalog />
		</Page>
	);
};
