import { ClientShell } from '@widgets/client-shell';
import { Outlet } from 'react-router';

export const ClientLayout = () => (
	<ClientShell>
		<Outlet />
	</ClientShell>
);
