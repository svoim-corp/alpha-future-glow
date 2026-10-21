import { BankHeader } from '@widgets/bank-header';
import { Outlet } from 'react-router';

import styles from './BankLayout.module.scss';

export const BankLayout = () => (
	<>
		<BankHeader />
		<div className={styles.content}>
			<Outlet />
		</div>
	</>
);
