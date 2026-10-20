import { ClientHeader } from '@widgets/client-header';
import { Outlet } from 'react-router';

export const ClientLayout = () => (
	<>
		<ClientHeader />
		<Outlet />
	</>
);
