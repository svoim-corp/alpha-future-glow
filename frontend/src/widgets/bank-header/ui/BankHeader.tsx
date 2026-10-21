import { useGetMeQuery } from '@entities/session';
import { SignOutButton } from '@features/sign-out';
import { ROUTES } from '@shared/config';
import { useTheme } from '@shared/lib/theme';
import { Icon } from '@shared/ui';
import { Link } from 'react-router';

import styles from './BankHeader.module.scss';

export const BankHeader = () => {
	const { data: user } = useGetMeQuery();
	const { theme, toggle } = useTheme();

	return (
		<header className={styles.header}>
			<div className={styles.inner}>
				<Link to={ROUTES.bankQueue} className={styles.brand}>
					Кредитный кабинет
					<span className={styles.role}>РАБОЧЕЕ МЕСТО БАНКА</span>
				</Link>
				<div className={styles.side}>
					{user && <span className={styles.user}>{user.fullName}, {user.position}</span>}
					<button
						type="button"
						className={styles.theme}
						aria-label="Тёмная тема"
						aria-pressed={theme === 'dark'}
						onClick={toggle}
					>
						<Icon name={theme === 'dark' ? 'sun' : 'moon'} />
					</button>
					<SignOutButton />
				</div>
			</div>
		</header>
	);
};
