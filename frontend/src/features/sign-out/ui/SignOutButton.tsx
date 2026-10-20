import { useLogoutMutation } from '@entities/session';
import { ROUTES } from '@shared/config';
import { Button } from '@shared/ui';
import { useNavigate } from 'react-router';

export const SignOutButton = () => {
	const navigate = useNavigate();
	const [logout, { isLoading }] = useLogoutMutation();

	const handleClick = async () => {
		try {
			await logout().unwrap();
		} finally {
			await navigate(ROUTES.login);
		}
	};

	return <Button variant="ghost" compact loading={isLoading} onClick={() => void handleClick()}>Выйти</Button>;
};
