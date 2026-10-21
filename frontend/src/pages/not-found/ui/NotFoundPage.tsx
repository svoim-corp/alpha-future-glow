import { ROUTES } from '@shared/config';
import { ButtonLink } from '@shared/ui';

import styles from './NotFoundPage.module.scss';

export const NotFoundPage = () => (
	<main className={styles.page}>
		<p className={styles.code} aria-hidden="true">404</p>
		<h1 className={styles.title}>Страница не найдена</h1>
		<p className={styles.text}>Возможно, ссылка устарела или адрес введён с ошибкой.</p>
		<ButtonLink to={ROUTES.home}>На главную</ButtonLink>
	</main>
);
