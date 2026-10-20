import { NotificationBell } from '@entities/notification';
import { useGetMeQuery } from '@entities/session';
import { SignOutButton } from '@features/sign-out';
import { ROUTES } from '@shared/config';
import { Button } from '@shared/ui';
import clsx from 'clsx';
import { type ReactNode, useEffect, useState } from 'react';
import { Link, NavLink, useLocation } from 'react-router';

import { useTheme } from '../model/useTheme';
import styles from './ClientShell.module.scss';

const NAV = [
	{ to: ROUTES.home, label: 'Главная', end: true },
	{ to: ROUTES.planner, label: 'Планер', end: false },
	{ to: ROUTES.applications, label: 'Мои заявки', end: true },
	{ to: ROUTES.credits, label: 'Мои кредиты', end: false },
	{ to: ROUTES.products, label: 'Каталог', end: false },
	{ to: ROUTES.documents, label: 'Документы', end: false },
	{ to: ROUTES.support, label: 'Поддержка', end: false },
] as const;

const linkClass = ({ isActive }: { isActive: boolean }) => clsx(styles.link, isActive && styles.active);

const initials = (fullName: string) => fullName
	.split(' ')
	.filter(Boolean)
	.slice(0, 2)
	.map((part) => part[0]?.toUpperCase())
	.join('');

export interface ClientShellProps {
	/** Содержимое страницы (слот, чтобы оболочка не зависела от роутера) */
	children: ReactNode;
}

/** Оболочка клиента: боковая навигация, верхняя панель с темой и профилем, мобильное меню */
export const ClientShell = ({ children }: ClientShellProps) => {
	const { data: user } = useGetMeQuery();
	const { theme, toggle } = useTheme();
	const { pathname } = useLocation();
	const [menuPath, setMenuPath] = useState<string | null>(null);
	const open = menuPath === pathname;
	const [profileOpen, setProfileOpen] = useState(false);

	useEffect(() => {
		if (!open && !profileOpen) {
			return undefined;
		}

		const onKey = (event: KeyboardEvent) => {
			if (event.key === 'Escape') {
				setMenuPath(null);
				setProfileOpen(false);
			}
		};
		document.addEventListener('keydown', onKey);

		return () => document.removeEventListener('keydown', onKey);
	}, [open, profileOpen]);

	const company = user?.company;

	return (
		<div className={styles.shell}>
			{open && (
				<button
					type="button"
					className={styles.backdrop}
					aria-label="Закрыть навигацию"
					onClick={() => setMenuPath(null)}
				/>
			)}
			<aside id="sidebar" className={clsx(styles.sidebar, open && styles.open)}>
				<Link to={ROUTES.home} className={styles.brand}>
					<span className={styles.mark} aria-hidden="true">к.</span>
					<span>Кредитный кабинет<small className={styles.sub}>ДЛЯ ВАШЕГО БИЗНЕСА</small></span>
				</Link>
				{company && (
					<div className={styles.company}>
						<strong>{company.name}</strong>
						<small>ИНН {company.inn}</small>
					</div>
				)}
				<nav className={styles.nav} aria-label="Разделы кабинета">
					{NAV.map((item) => (
						<NavLink key={item.to} to={item.to} end={item.end} className={linkClass}>{item.label}</NavLink>
					))}
				</nav>
				<div className={styles.bottom}>
					<SignOutButton />
				</div>
			</aside>
			<div className={styles.content}>
				<header className={styles.topbar}>
					<Button
						variant="secondary"
						compact
						className={styles.menuButton}
						aria-expanded={open}
						aria-controls="sidebar"
						onClick={() => setMenuPath(open ? null : pathname)}
					>
						Меню
					</Button>
					<div className={styles.actions}>
						<Button variant="ghost" compact aria-pressed={theme === 'dark'} onClick={toggle}>
							{theme === 'dark' ? 'Светлая тема' : 'Тёмная тема'}
						</Button>
						<NotificationBell />
						{user && (
							<div className={styles.profile}>
								<Button
									variant="ghost"
									compact
									aria-expanded={profileOpen}
									aria-controls="profile-panel"
									onClick={() => setProfileOpen(!profileOpen)}
								>
									<span className={styles.avatar} aria-hidden="true">{initials(user.fullName)}</span>
									{user.fullName}
								</Button>
								{profileOpen && (
									<div id="profile-panel" className={styles.popover}>
										<strong>{user.fullName}</strong>
										<span>{user.position}</span>
										{company && <span>{company.name}, ИНН {company.inn}</span>}
									</div>
								)}
							</div>
						)}
					</div>
				</header>
				{children}
			</div>
		</div>
	);
};
