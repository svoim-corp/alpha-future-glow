import { type ApiError } from './types';

const isRecord = (value: unknown): value is Record<string, unknown> => typeof value === 'object' && value !== null;

const isApiError = (value: unknown): value is ApiError => (
	isRecord(value) && typeof value.code === 'string' && typeof value.message === 'string'
);

/** Достаёт единый формат ошибки бэкенда (раздел 3 API_SPEC) из ошибки RTK Query */
export const getApiError = (error: unknown): ApiError | undefined => (
	isRecord(error) && isApiError(error.data) ? error.data : undefined
);

/** HTTP-статус ошибки RTK Query; у сетевых сбоев его нет */
export const getErrorStatus = (error: unknown): number | undefined => (
	isRecord(error) && typeof error.status === 'number' ? error.status : undefined
);

/** Сообщение бэкенда или запасной текст, если ответа с телом не было */
export const getErrorMessage = (error: unknown, fallback: string): string => getApiError(error)?.message ?? fallback;

/** Ошибки полей `{ поле: сообщение }` из VALIDATION_ERROR */
export const getFieldErrors = (error: unknown): Record<string, string> => {
	const fields = getApiError(error)?.fields ?? [];

	return Object.fromEntries(fields.map((field) => [field.field, field.message]));
};
