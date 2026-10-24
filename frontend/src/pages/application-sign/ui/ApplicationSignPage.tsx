import { paths } from '@shared/config';
import { Navigate, useParams } from 'react-router';

/** Подписание живёт на странице заявки; адрес оставлен для старых ссылок */
export const ApplicationSignPage = () => {
	const { id = '' } = useParams();

	return <Navigate to={paths.application(id)} replace />;
};
