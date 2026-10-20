import { z } from 'zod';

const money = (label: string) => z.string().trim().regex(/^\d+$/u, `${label}: введите целое число рублей`);

export const paramsSchema = z.object({
	amount: money('Сумма').refine((value) => Number(value) > 0, 'Сумма должна быть больше нуля'),
	term: z.string().trim().regex(/^\d+$/u, 'Срок: введите число месяцев')
		.refine((value) => Number(value) > 0, 'Срок должен быть больше нуля'),
	graceMonths: z.string().trim().regex(/^\d*$/u, 'Отсрочка: введите число месяцев'),
	overdraftLimit: z.string().trim().regex(/^\d*$/u, 'Лимит: введите целое число рублей'),
	purposeText: z.string().trim().max(500, 'Не больше 500 символов'),
});

export type ParamsValues = z.infer<typeof paramsSchema>;
