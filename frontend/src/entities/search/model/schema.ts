import { z } from 'zod';

export const searchKindSchema = z.enum(['applications', 'loans', 'plans', 'products', 'faq', 'tickets']);
const itemSchema = z.object({
	id: z.string().min(1).max(128).regex(/^[A-Za-z0-9_-]+$/u),
	title: z.string(),
	snippet: z.string(),
	state: z.string(),
	step: z.string(),
}).strict();
export const searchResponseSchema = z.object({
	query: z.string(),
	groups: z.array(z.object({
		type: searchKindSchema,
		items: z.array(itemSchema).max(20),
		hasMore: z.boolean(),
		nextCursor: z.string().nullable(),
		unavailable: z.boolean(),
	}).strict()).min(1).max(6).refine((groups) => new Set(groups.map((group) => group.type)).size === groups.length),
}).strict();

export type SearchKind = z.infer<typeof searchKindSchema>;
export type SearchItem = z.infer<typeof itemSchema>;
export type SearchResponse = z.infer<typeof searchResponseSchema>;
export interface SearchArgs {
	query: string;
	type?: SearchKind;
	limit: number;
	cursor?: string;
	/** Только для изоляции кеша. Сервер определяет доступ по сессии. */
	userId: string;
	companyId: string;
}
