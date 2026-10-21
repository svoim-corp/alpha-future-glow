import { Page } from '@shared/ui';
import { ProductCatalog } from '@widgets/product-catalog';

export const ProductsPage = () => (
	<Page
		title="Кредиты для бизнеса"
		subtitle="Расскажите, на что нужны деньги, — подскажем подходящий продукт."
	>
		<ProductCatalog />
	</Page>
);
