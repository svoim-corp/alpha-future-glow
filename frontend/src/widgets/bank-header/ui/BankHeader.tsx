import { useGetMeQuery } from '@entities/session';
import { SignOutButton } from '@features/sign-out';
import { ROUTES } from '@shared/config';
import { Link } from 'react-router';

import styles from './BankHeader.module.scss';

export const BankHeader = () => {
	const { data: user } = useGetMeQuery();

	return (
		<header className={styles.header}>
			<div className={styles.inner}>
				<Link to={ROUTES.bankQueue} className={styles.brand}>
					Кредитный кабинет
					<span className={styles.role}>РАБОЧЕЕ МЕСТО БАНКА</span>
				</Link>
				<div className={styles.side}>
					{user && <span className={styles.user}>{user.fullName}, {user.position}</span>}
					<SignOutButton />
				</div>
			</div>
		</header>
	);
};
