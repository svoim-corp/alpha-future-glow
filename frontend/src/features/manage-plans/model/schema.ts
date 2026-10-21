import { z } from 'zod';

export const planSchema = z.object({
	title: z.string().trim().min(2, 'Назовите план'),
	budget: z.string().trim().regex(/^\d+$/u, { error: 'Бюджет — целое число рублей', abort: true })
		.refine((value) => Number(value) > 0, 'Бюджет должен быть больше нуля'),
	goal: z.string(),
	dateMode: z.enum(['UNSCHEDULED', 'DAY', 'PERIOD']),
	startDate: z.string(),
	endDate: z.string(),
	note: z.string().trim().max(500, 'Не больше 500 символов'),
	email: z.boolean(),
}).superRefine((value, ctx) => {
	if (value.dateMode !== 'UNSCHEDULED' && value.startDate === '') {
		ctx.addIssue({ code: 'custom', path: ['startDate'], message: 'Укажите дату начала' });
	}

	if (value.dateMode === 'PERIOD') {
		if (value.endDate === '') {
			ctx.addIssue({ code: 'custom', path: ['endDate'], message: 'Укажите дату окончания' });
		} else if (value.startDate !== '' && value.endDate < value.startDate) {
			ctx.addIssue({ code: 'custom', path: ['endDate'], message: 'Окончание не может быть раньше начала' });
		}
	}
});

export type PlanValues = z.infer<typeof planSchema>;
