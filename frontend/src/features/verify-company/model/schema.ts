import { z } from 'zod';

/** ИНН: 10 цифр у юрлица, 12 у ИП. Контрольную сумму проверяет сервер */
export const innSchema = z.object({
	inn: z.string().trim().regex(/^(\d{10}|\d{12})$/u, 'ИНН — 10 цифр для организации или 12 для ИП'),
});

export const mchdSchema = z.object({
	mchdNumber: z.string().trim().min(1, 'Введите номер доверенности'),
});

export type InnValues = z.infer<typeof innSchema>;
export type MchdValues = z.infer<typeof mchdSchema>;
