import { homeRouteFor, useGetMeQuery } from '@entities/session';
import { SignInForm } from '@features/sign-in';
import { Card } from '@shared/ui';
import { Navigate } from 'react-router';

import styles from './LoginPage.module.scss';

export const LoginPage = () => {
	const { data: user } = useGetMeQuery();

	if (user) {
		return <Navigate to={homeRouteFor(user.role)} replace />;
	}

	return (
		<main className={styles.page}>
			<div className={styles.panel}>
				<div className={styles.head}>
					<span className={styles.logo} aria-hidden="true">К.</span>
					<h1 className={styles.title}>Умный кредитный кабинет</h1>
					<p className={styles.lead}>Войдите, чтобы подобрать кредит и следить за заявкой.</p>
				</div>
				<Card>
					<SignInForm />
				</Card>
				<Card aria-labelledby="demo-accounts">
					<h2 id="demo-accounts" className={styles.demoTitle}>Демо-учётки (пароль у всех demo123)</h2>
					<ul className={styles.demo}>
						<li>smirnov — клиент, генеральный директор</li>
						<li>petrova — клиент, главный бухгалтер</li>
						<li>ivanova — сотрудник банка</li>
					</ul>
				</Card>
			</div>
		</main>
	);
};
