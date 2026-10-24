import { NotificationBell } from '@entities/notification';
import { useGetMeQuery } from '@entities/session';
import { GlobalSearch } from '@features/global-search';
import { SignOutButton } from '@features/sign-out';
import { ROUTES } from '@shared/config';
import { useTheme } from '@shared/lib/theme';
import { Icon, type IconName } from '@shared/ui';
import clsx from 'clsx';
import { type ReactNode, useEffect, useRef, useState } from 'react';
import { Link, NavLink, useLocation } from 'react-router';

import styles from './ClientShell.module.scss';

interface NavEntry {
	to: string;
	label: string;
	icon: IconName;
	end: boolean;
	keywords?: readonly string[];
}

const NAV: NavEntry[] = [
	{ to: ROUTES.home, label: 'Главная', icon: 'home', end: true, keywords: ['обзор', 'виджеты'] },
	{ to: ROUTES.planner, label: 'Планер', icon: 'calendar', end: false, keywords: ['план', 'идея'] },
	{ to: ROUTES.applications, label: 'Мои заявки', icon: 'applications', end: true,
		keywords: ['заявка', 'черновик', 'оформление'] },
	{ to: ROUTES.credits, label: 'Мои кредиты', icon: 'loans', end: false,
		keywords: ['кредит', 'договор', 'платёж', 'график'] },
	{ to: ROUTES.products, label: 'Каталог', icon: 'catalog', end: false, keywords: ['продукт', 'финансирование'] },
	{ to: ROUTES.support, label: 'Поддержка', icon: 'support', end: false,
		keywords: ['вопрос', 'ответ', 'обращение', 'помощь'] },
];

const initials = (fullName: string) => {
	// ФИО в формате «Фамилия Имя Отчество»: инициалы — имя и фамилия, как на макете («ОС» для Олега Смирнова)
	const [last, first] = fullName.split(' ').filter(Boolean);

	return `${first?.[0] ?? ''}${last?.[0] ?? ''}`.toUpperCase();
};

const shortName = (fullName: string) => {
	const [last, first] = fullName.split(' ').filter(Boolean);

	return first && last ? `${first} ${last}` : fullName;
};

type Popover = 'profile' | null;

export interface ClientShellProps {
	/** Содержимое страницы (слот, чтобы оболочка не зависела от роутера) */
	children: ReactNode;
}

/** Оболочка клиента по макету shell.html: боковая навигация, верхняя панель, мобильное меню */
export const ClientShell = ({ children }: ClientShellProps) => {
	const { data: user } = useGetMeQuery();
	const { theme, toggle } = useTheme();
	const { pathname } = useLocation();
	const [menuPath, setMenuPath] = useState<string | null>(null);
	const [popover, setPopover] = useState<Popover>(null);
	const topbarRef = useRef<HTMLElement>(null);
	const open = menuPath === pathname;
	const company = user?.company;

	useEffect(() => {
		if (!open && !popover) {
			return undefined;
		}

		const onKey = (event: KeyboardEvent) => {
			if (event.key === 'Escape') {
				setMenuPath(null);
				setPopover(null);
			}
		};
		const onClick = (event: MouseEvent) => {
			if (!topbarRef.current?.contains(event.target as Node)) {
				setPopover(null);
			}
		};
		document.addEventListener('keydown', onKey);
		document.addEventListener('mousedown', onClick);

		return () => {
			document.removeEventListener('keydown', onKey);
			document.removeEventListener('mousedown', onClick);
		};
	}, [open, popover]);

	const toggleMenu = () => setMenuPath(open ? null : pathname);
	const togglePopover = (name: Exclude<Popover, null>) => setPopover(popover === name ? null : name);

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
					<span>
						Кредитный<br />кабинет
						<span className={styles.sub}>ДЛЯ ВАШЕГО БИЗНЕСА</span>
					</span>
				</Link>
				<button type="button" className={styles.close} aria-label="Закрыть навигацию" onClick={() => setMenuPath(null)}>
					<Icon name="close" />
				</button>
				<span className={styles.caption}>ВАШ КАБИНЕТ</span>
				<nav className={styles.nav} aria-label="Разделы кабинета">
					{NAV.map((item) => (
						<NavLink
							key={item.to}
							to={item.to}
							end={item.end}
							className={({ isActive }) => clsx(styles.item, isActive && styles.active)}
						>
							<Icon name={item.icon} size={20} />
							<span>{item.label}</span>
							<span className={styles.dot} aria-hidden="true" />
						</NavLink>
					))}
				</nav>
				<div className={styles.bottom}>
					<SignOutButton className={clsx(styles.item, styles.logout)}>
						<Icon name="logout" size={20} />
						<span>Выйти</span>
					</SignOutButton>
				</div>
			</aside>
			<div className={styles.workspace}>
				<header className={styles.topbar} ref={topbarRef}>
					<div className={styles.searchWrap}>
						<GlobalSearch sections={NAV} onOpen={() => setPopover(null)} />
					</div>
					<div className={styles.actions}>
						<NotificationBell />
						<button
							type="button"
							className={styles.iconButton}
							aria-label="Тёмная тема"
							aria-pressed={theme === 'dark'}
							onClick={toggle}
						>
							<Icon name={theme === 'dark' ? 'sun' : 'moon'} />
						</button>
						<span className={styles.divider} />
						{user && (
							<div className={styles.dropdown}>
								<button
									type="button"
									className={styles.profile}
									aria-expanded={popover === 'profile'}
									aria-controls="profile-panel"
									onClick={() => togglePopover('profile')}
								>
									<span className={styles.avatar} aria-hidden="true">{initials(user.fullName)}</span>
									<span className={styles.profileText}>
										<strong>{shortName(user.fullName)}</strong>
										<small>{user.position}</small>
									</span>
									<Icon name="chevron" size={16} />
								</button>
								{popover === 'profile' && (
									<div id="profile-panel" className={clsx(styles.popover, styles.profilePanel)}>
										<span className={styles.caption2}>ПРОФИЛЬ</span>
										<strong>{user.fullName}</strong>
										<p>{user.position}</p>
										{company && (
											<div className={styles.profileCompany}>
												<Icon name="building" size={20} />
												<span>{company.name}<small>ИНН {company.inn}</small></span>
											</div>
										)}
									</div>
								)}
							</div>
						)}
						<button
							type="button"
							className={clsx(styles.iconButton, styles.menuButton)}
							aria-label="Открыть навигацию"
							aria-expanded={open}
							aria-controls="sidebar"
							onClick={toggleMenu}
						>
							<Icon name="menu" />
						</button>
					</div>
				</header>
				<div className={styles.canvas}>{children}</div>
			</div>
		</div>
	);
};
