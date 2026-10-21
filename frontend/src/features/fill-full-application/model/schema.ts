import { z } from 'zod';

export const stage2Schema = z.object({
	sameAsRegistration: z.boolean(),
	address: z.string().trim().max(300, 'Не больше 300 символов'),
	guarantorName: z.string().trim().max(200, 'Не больше 200 символов'),
	guarantorInn: z.string().trim().regex(/^(\d{12})?$/u, 'ИНН физлица — 12 цифр'),
}).superRefine((value, ctx) => {
	if (!value.sameAsRegistration && value.address === '') {
		ctx.addIssue({ code: 'custom', path: ['address'], message: 'Укажите фактический адрес' });
	}
});

export type Stage2Values = z.infer<typeof stage2Schema>;
