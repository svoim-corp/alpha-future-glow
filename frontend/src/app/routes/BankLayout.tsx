import { BankShell } from '@widgets/bank-shell';
import { Outlet } from 'react-router';

export const BankLayout = () => (
	<BankShell>
		<Outlet />
	</BankShell>
);
