import { getApiError, getErrorMessage, getErrorStatus, getFieldErrors } from './errors';

const validation = {
	status: 400,
	data: {
		code: 'VALIDATION_ERROR',
		message: 'Не все поля заполнены',
		fields: [{ field: 'guarantor.inn', code: 'REQUIRED', message: 'Укажите ИНН поручителя' }],
	},
};

describe('getApiError', () => {
	it('возвращает тело ошибки бэкенда', () => {
		expect(getApiError(validation)?.code).toBe('VALIDATION_ERROR');
	});

	it.each([[undefined], [null], ['строка'], [{ status: 500 }], [{ data: { code: 1 } }]])(
		'не принимает чужую форму %j',
		(value) => {
			expect(getApiError(value)).toBeUndefined();
		},
	);
});

describe('getErrorStatus', () => {
	it('возвращает HTTP-статус, у сетевой ошибки его нет', () => {
		expect(getErrorStatus(validation)).toBe(400);
		expect(getErrorStatus({ status: 'FETCH_ERROR' })).toBeUndefined();
		expect(getErrorStatus(null)).toBeUndefined();
	});
});

describe('getErrorMessage', () => {
	it('берёт сообщение сервера, иначе запасной текст', () => {
		expect(getErrorMessage(validation, 'Сбой')).toBe('Не все поля заполнены');
		expect(getErrorMessage({ status: 500 }, 'Сбой')).toBe('Сбой');
	});
});

describe('getFieldErrors', () => {
	it('собирает ошибки по полям, без полей — пустой объект', () => {
		expect(getFieldErrors(validation)).toEqual({ 'guarantor.inn': 'Укажите ИНН поручителя' });
		expect(getFieldErrors(undefined)).toEqual({});
	});
});
