import { homeRouteFor, useGetMeQuery, type UserRole } from '@entities/session';
import { getErrorStatus } from '@shared/api';
import { ROUTES } from '@shared/config';
import { AsyncBoundary } from '@shared/ui';
import { Navigate, Outlet } from 'react-router';

interface RequireRoleProps {
	allow: UserRole;
}

/** Пускает на страницы только нужную роль: без сессии — на вход, чужая роль — на свою главную */
export const RequireRole = ({ allow }: RequireRoleProps) => {
	const { data: user, isLoading, isError, error, refetch } = useGetMeQuery(undefined, {
		refetchOnMountOrArgChange: true,
	});

	if (!isLoading && isError && getErrorStatus(error) === 401) {
		return <Navigate to={ROUTES.login} replace />;
	}

	if (user && user.role !== allow) {
		return <Navigate to={homeRouteFor(user.role)} replace />;
	}

	return (
		<AsyncBoundary
			isLoading={isLoading}
			isError={isError}
			onRetry={() => void refetch()}
			errorText="Не удалось проверить доступ."
		>
			<Outlet />
		</AsyncBoundary>
	);
};
