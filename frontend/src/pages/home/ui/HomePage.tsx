import { useGetMeQuery } from '@entities/session';
import { HomeDashboard } from '@widgets/home-dashboard';

import { firstName, greeting } from '../model/greeting';
import styles from './HomePage.module.scss';

const todayFormat = new Intl.DateTimeFormat('ru-RU', { day: 'numeric', month: 'long', year: 'numeric' });

export const HomePage = () => {
	const { data: user } = useGetMeQuery();
	const now = new Date();

	return (
		<main>
			<div className={styles.heading}>
				<span className={styles.eyebrow}>ВАШ БИЗНЕС — В ЦЕНТРЕ ВНИМАНИЯ</span>
				<h1>{user ? `${greeting(now.getHours())}, ${firstName(user.fullName)}` : 'Добро пожаловать'}</h1>
				<p>Всё важное о финансировании — перед вами.</p>
			</div>
			<div className={styles.meta}>
				<span><i />{user?.company?.name ?? 'Ваша компания'}</span>
				<span>{todayFormat.format(now)}</span>
			</div>
			<HomeDashboard />
		</main>
	);
};
