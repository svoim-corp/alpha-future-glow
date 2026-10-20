import { ROUTES } from '@shared/config';
import { Link } from 'react-router';

import styles from './NotFoundPage.module.scss';

export const NotFoundPage = () => (
	<main className={styles.page}>
		<h1>Страница не найдена</h1>
		<Link to={ROUTES.home}>На главную</Link>
	</main>
);
