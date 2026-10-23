import { useGetMeQuery } from '@entities/session';
import { HomeDashboard } from '@widgets/home-dashboard';

import { firstName, greeting } from '../model/greeting';
import styles from './HomePage.module.scss';

export const HomePage = () => {
	const { data: user } = useGetMeQuery();
	const now = new Date();

	return (
		<main>
			<div className={styles.heading}>
				<h1>{user ? `${greeting(now.getHours())}, ${firstName(user.fullName)}` : 'Добро пожаловать'}</h1>
				<p>Всё важное о финансировании — перед вами.</p>
			</div>
			<HomeDashboard />
		</main>
	);
};
