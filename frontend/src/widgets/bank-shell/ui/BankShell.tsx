import { useGetMeQuery } from '@entities/session';
import { SignOutButton } from '@features/sign-out';
import { ROUTES } from '@shared/config';
import { useTheme } from '@shared/lib/theme';
import { Icon } from '@shared/ui';
import clsx from 'clsx';
import { type ReactNode } from 'react';
import { Link } from 'react-router';

import styles from './BankShell.module.scss';

const initials = (fullName: string) => {
	// ФИО в формате «Фамилия Имя Отчество»: инициалы — имя и фамилия, как в оболочке клиента
	const [last, first] = fullName.split(' ').filter(Boolean);

	return `${first?.[0] ?? ''}${last?.[0] ?? ''}`.toUpperCase();
};

const shortName = (fullName: string) => {
	const [last, first] = fullName.split(' ').filter(Boolean);

	return first && last ? `${first} ${last}` : fullName;
};

export interface BankShellProps {
	/** Содержимое страницы (слот, чтобы оболочка не зависела от роутера) */
	children: ReactNode;
}

/** Оболочка рабочего места сотрудника: верхняя панель в стиле клиентской, без боковой навигации */
export const BankShell = ({ children }: BankShellProps) => {
	const { data: user } = useGetMeQuery();
	const { theme, toggle } = useTheme();

	return (
		<div className={styles.shell}>
			<header className={styles.topbar}>
				<Link to={ROUTES.bankQueue} className={styles.brand}>
					<span className={styles.mark} aria-hidden="true">к.</span>
					<span className={styles.name}>
						Кредитный<br />кабинет
						<span className={styles.sub}>РАБОЧЕЕ МЕСТО БАНКА</span>
					</span>
				</Link>
				<div className={styles.actions}>
					<button
						type="button"
						className={styles.iconButton}
						aria-label="Тёмная тема"
						aria-pressed={theme === 'dark'}
						onClick={toggle}
					>
						<Icon name={theme === 'dark' ? 'sun' : 'moon'} />
					</button>
					{user && (
						<>
							<span className={styles.divider} />
							<div className={styles.profile}>
								<span className={styles.avatar} aria-hidden="true">{initials(user.fullName)}</span>
								<span className={styles.profileText}>
									<strong>{shortName(user.fullName)}</strong>
									<small>{user.position}</small>
								</span>
							</div>
						</>
					)}
					<SignOutButton className={clsx(styles.iconButton, styles.logout)}>
						<Icon name="logout" size={20} />
						<span className="visually-hidden">Выйти</span>
					</SignOutButton>
				</div>
			</header>
			<div className={styles.canvas}>{children}</div>
		</div>
	);
};
