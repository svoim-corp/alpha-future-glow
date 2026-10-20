import { BankHeader } from '@widgets/bank-header';
import { Outlet } from 'react-router';

export const BankLayout = () => (
	<>
		<BankHeader />
		<Outlet />
	</>
);
