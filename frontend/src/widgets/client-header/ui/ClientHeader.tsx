import { NotificationBell } from '@entities/notification';
import { useGetMeQuery } from '@entities/session';
import { SignOutButton } from '@features/sign-out';
import { ROUTES } from '@shared/config';
import clsx from 'clsx';
import { Link, NavLink } from 'react-router';

import styles from './ClientHeader.module.scss';

const NAV = [
	{ to: ROUTES.home, label: 'Главная', end: true },
	{ to: ROUTES.products, label: 'Каталог', end: false },
	{ to: ROUTES.credits, label: 'Мои кредиты', end: true },
	{ to: ROUTES.creditsCalendar, label: 'Платёжный календарь', end: false },
] as const;

const linkClass = ({ isActive }: { isActive: boolean }) => clsx(styles.link, isActive && styles.active);

export const ClientHeader = () => {
	const { data: user } = useGetMeQuery();

	return (
		<header className={styles.header}>
			<div className={styles.inner}>
				<Link to={ROUTES.home} className={styles.brand}>Кредитный кабинет</Link>
				<nav className={styles.nav} aria-label="Основная навигация">
					{NAV.map((item) => (
						<NavLink key={item.to} to={item.to} end={item.end} className={linkClass}>{item.label}</NavLink>
					))}
				</nav>
				<div className={styles.side}>
					<NavLink to={ROUTES.support} className={linkClass}>Поддержка</NavLink>
					<NotificationBell />
					{user && <span className={styles.user}>{user.fullName}</span>}
					<SignOutButton />
				</div>
			</div>
		</header>
	);
};
