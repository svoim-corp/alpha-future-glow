import { signInSchema } from './schema';

describe('signInSchema', () => {
	it('принимает логин и пароль', () => {
		expect(signInSchema.safeParse({ login: 'smirnov', password: 'demo123' }).success).toBe(true);
	});

	it('требует логин и пароль, пробелы в логине не считаются вводом', () => {
		const result = signInSchema.safeParse({ login: '   ', password: '' });

		expect(result.success).toBe(false);
		expect(result.error?.issues.map((issue) => issue.message)).toEqual(['Введите логин', 'Введите пароль']);
	});
});
