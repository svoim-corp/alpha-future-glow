import { NotificationBell } from '@entities/notification';
import { useGetMeQuery } from '@entities/session';
import { SignOutButton } from '@features/sign-out';
import { ROUTES } from '@shared/config';
import { Icon, type IconName } from '@shared/ui';
import clsx from 'clsx';
import { type ReactNode, useEffect, useRef, useState } from 'react';
import { Link, NavLink, useLocation, useNavigate } from 'react-router';

import { useTheme } from '../model/useTheme';
import styles from './ClientShell.module.scss';

interface NavEntry {
	to: string;
	label: string;
	icon: IconName;
	end: boolean;
}

const NAV: NavEntry[] = [
	{ to: ROUTES.home, label: 'Главная', icon: 'home', end: true },
	{ to: ROUTES.planner, label: 'Планер', icon: 'calendar', end: false },
	{ to: ROUTES.applications, label: 'Мои заявки', icon: 'applications', end: true },
	{ to: ROUTES.credits, label: 'Мои кредиты', icon: 'loans', end: false },
	{ to: ROUTES.products, label: 'Каталог', icon: 'catalog', end: false },
	{ to: ROUTES.documents, label: 'Документы', icon: 'documents', end: false },
	{ to: ROUTES.support, label: 'Поддержка', icon: 'support', end: false },
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

/** «ООО «Северная пекарня»» → «Северная пекарня»: организационная форма в карточке лишняя */
const companyTitle = (name: string) => name.match(/«(.+)»/)?.[1] ?? name;

const companyInitials = (name: string) => companyTitle(name)
	.split(' ')
	.slice(0, 2)
	.map((word) => word[0]?.toUpperCase())
	.join('');

type Popover = 'language' | 'profile' | 'search' | null;

export interface ClientShellProps {
	/** Содержимое страницы (слот, чтобы оболочка не зависела от роутера) */
	children: ReactNode;
}

/** Оболочка клиента по макету shell.html: боковая навигация, верхняя панель, мобильное меню */
export const ClientShell = ({ children }: ClientShellProps) => {
	const { data: user } = useGetMeQuery();
	const { theme, toggle } = useTheme();
	const { pathname } = useLocation();
	const navigate = useNavigate();
	const [menuPath, setMenuPath] = useState<string | null>(null);
	const [popover, setPopover] = useState<Popover>(null);
	const [query, setQuery] = useState('');
	const topbarRef = useRef<HTMLElement>(null);
	const open = menuPath === pathname;
	const company = user?.company;
	const found = NAV.filter((item) => item.label.toLowerCase().includes(query.trim().toLowerCase()));

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
	const goTo = (to: string) => {
		setPopover(null);
		setQuery('');
		void navigate(to);
	};

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
				{company && (
					<div className={styles.company}>
						<span className={styles.companyAvatar} aria-hidden="true">{companyInitials(company.name)}</span>
						<div>
							<strong>{companyTitle(company.name)}</strong>
							<small>ИНН {company.inn}</small>
						</div>
					</div>
				)}
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
						<label className={styles.search}>
							<Icon name="search" />
							<span className="visually-hidden">Поиск по кабинету</span>
							<input
								type="search"
								placeholder="Поиск по кабинету"
								value={query}
								autoComplete="off"
								onFocus={() => setPopover('search')}
								onChange={(event) => setQuery(event.target.value)}
								onKeyDown={(event) => {
									if (event.key === 'Enter' && found[0]) {
										goTo(found[0].to);
									}
								}}
							/>
						</label>
						{popover === 'search' && (
							<div className={clsx(styles.popover, styles.searchPanel)}>
								<span className={styles.caption2}>РАЗДЕЛЫ КАБИНЕТА</span>
								{found.length === 0 && <p className={styles.hint}>Ничего не найдено</p>}
								<ul className={styles.results}>
									{found.map((item) => (
										<li key={item.to}>
											<button type="button" onClick={() => goTo(item.to)}>
												<Icon name={item.icon} size={18} />
												{item.label}
											</button>
										</li>
									))}
								</ul>
							</div>
						)}
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
						<div className={styles.dropdown}>
							<button
								type="button"
								className={styles.language}
								aria-expanded={popover === 'language'}
								aria-controls="language-panel"
								onClick={() => togglePopover('language')}
							>
								<Icon name="globe" size={20} />
								<span>RU</span>
								<Icon name="chevron" size={16} />
							</button>
							{popover === 'language' && (
								<div id="language-panel" className={clsx(styles.popover, styles.languagePanel)}>
									<span className={styles.caption2}>ЯЗЫК ИНТЕРФЕЙСА</span>
									<div className={styles.option}>
										Русский <span className={styles.tag}>RU</span>
										<Icon name="check" size={17} className={styles.check} />
									</div>
									<div className={clsx(styles.option, styles.muted)}>
										English <span className={styles.later}>Позже</span>
									</div>
								</div>
							)}
						</div>
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
