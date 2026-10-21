import { useLoginMutation } from '@entities/session';
import { zodResolver } from '@hookform/resolvers/zod';
import { getErrorStatus } from '@shared/api';
import { Alert, Button, Field } from '@shared/ui';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { useNavigate } from 'react-router';

import { signInSchema,type SignInValues } from '../model/schema';
import styles from './SignInForm.module.scss';

export const SignInForm = () => {
	const navigate = useNavigate();
	const [login, { isLoading }] = useLoginMutation();
	const [failure, setFailure] = useState<string | null>(null);
	const { register, handleSubmit, formState: { errors } } = useForm<SignInValues>({
		resolver: zodResolver(signInSchema),
		defaultValues: { login: '', password: '' },
	});

	const onSubmit = handleSubmit(async (values) => {
		setFailure(null);

		try {
			const response = await login(values).unwrap();
			await navigate(response.redirect);
		} catch (error) {
			setFailure(
				getErrorStatus(error) === 401
					? 'Неверный логин или пароль'
					: 'Не удалось войти. Проверьте соединение и повторите попытку.',
			);
		}
	});

	return (
		<form className={styles.form} onSubmit={(event) => void onSubmit(event)} noValidate>
			<Field label="Логин" autoComplete="username" error={errors.login?.message} {...register('login')} />
			<Field
				label="Пароль"
				type="password"
				autoComplete="current-password"
				error={errors.password?.message}
				{...register('password')}
			/>
			{failure && <Alert tone="danger" title={failure} />}
			<Button type="submit" loading={isLoading}>Войти</Button>
		</form>
	);
};
