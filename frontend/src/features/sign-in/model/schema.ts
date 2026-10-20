import { z } from 'zod';

export const signInSchema = z.object({
	login: z.string().trim().min(1, 'Введите логин'),
	password: z.string().min(1, 'Введите пароль'),
});

export type SignInValues = z.infer<typeof signInSchema>;
