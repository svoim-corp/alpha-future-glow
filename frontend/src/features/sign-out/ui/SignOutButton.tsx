import { useLogoutMutation } from '@entities/session';
import { ROUTES } from '@shared/config';
import { Button } from '@shared/ui';
import { type ReactNode } from 'react';
import { useNavigate } from 'react-router';

export interface SignOutButtonProps {
	/** Если задан className, кнопка рисуется простым <button> (пункт меню), а не кнопкой дизайн-системы */
	className?: string;
	children?: ReactNode;
}

export const SignOutButton = ({ className, children }: SignOutButtonProps) => {
	const navigate = useNavigate();
	const [logout, { isLoading }] = useLogoutMutation();

	const handleClick = async () => {
		try {
			await logout().unwrap();
		} finally {
			await navigate(ROUTES.login);
		}
	};

	if (className) {
		return (
			<button type="button" className={className} disabled={isLoading} onClick={() => void handleClick()}>
				{children ?? 'Выйти'}
			</button>
		);
	}

	return <Button variant="ghost" compact loading={isLoading} onClick={() => void handleClick()}>Выйти</Button>;
};
