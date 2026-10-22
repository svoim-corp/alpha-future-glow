import { z } from 'zod';

const validDate = (value: string): boolean => {
	if (!/^\d{4}-\d{2}-\d{2}$/u.test(value)) {
		return false;
	}

	const parsed = new Date(`${value}T00:00:00Z`);

	return Number.isFinite(parsed.getTime()) && parsed.toISOString().slice(0, 10) === value;
};

export const planSchema = z.object({
	title: z.string().trim().min(1, 'Назовите план').max(200, 'Не больше 200 символов'),
	budget: z.string().trim().regex(/^\d+$/u, { error: 'Бюджет — целое число рублей', abort: true })
		.refine((value) => Number(value) > 0, 'Бюджет должен быть больше нуля')
		.refine((value) => Number.isSafeInteger(Number(value)), 'Сумма слишком велика'),
	goal: z.string().trim().min(1, 'Опишите цель').max(64, 'Не больше 64 символов'),
	dateMode: z.enum(['UNSCHEDULED', 'DAY', 'PERIOD']),
	startDate: z.string(),
	endDate: z.string(),
	note: z.string().max(2000, 'Не больше 2000 символов'),
	inApp: z.boolean(),
}).superRefine((value, ctx) => {
	if (value.dateMode === 'UNSCHEDULED') {
		return;
	}

	if (value.startDate === '') {
		ctx.addIssue({ code: 'custom', path: ['startDate'], message: 'Укажите дату начала' });
	} else if (!validDate(value.startDate)) {
		ctx.addIssue({ code: 'custom', path: ['startDate'], message: 'Укажите корректную дату' });
	}

	if (value.dateMode === 'PERIOD') {
		if (value.endDate === '') {
			ctx.addIssue({ code: 'custom', path: ['endDate'], message: 'Укажите дату окончания' });
		} else if (!validDate(value.endDate)) {
			ctx.addIssue({ code: 'custom', path: ['endDate'], message: 'Укажите корректную дату' });
		} else if (value.startDate !== '' && value.endDate < value.startDate) {
			ctx.addIssue({ code: 'custom', path: ['endDate'], message: 'Окончание не может быть раньше начала' });
		}
	}
});

export type PlanValues = z.infer<typeof planSchema>;
