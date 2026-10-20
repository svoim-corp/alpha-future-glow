import { paths } from '@shared/config';
import { type Tone } from '@shared/ui';

import { type ApplicationShort, type ApplicationStatus } from './types';

export interface StatusView {
	label: string;
	tone: Tone;
}

/** Подписи клиенту и семантика статусов (словарь UI kit; API-коды не меняются) */
export const STATUS_VIEW: Record<ApplicationStatus, StatusView> = {
	DRAFT: { label: 'Черновик', tone: 'neutral' },
	SUBMITTED: { label: 'Заявка отправлена', tone: 'info' },
	DATA_CHECK: { label: 'Проверяем данные', tone: 'info' },
	INFO_REQUESTED: { label: 'Нужна информация', tone: 'warning' },
	SCORING: { label: 'Оцениваем заявку', tone: 'info' },
	ON_DECISION: { label: 'Принимаем решение', tone: 'info' },
	APPROVED: { label: 'Одобрено', tone: 'success' },
	COUNTER_OFFER: { label: 'Предложены другие условия', tone: 'warning' },
	REJECTED: { label: 'Не одобрено', tone: 'danger' },
	SIGNING: { label: 'Ожидает подписания', tone: 'warning' },
	ISSUED: { label: 'Кредит выдан', tone: 'success' },
	CLOSED: { label: 'Заявка закрыта', tone: 'neutral' },
};

/** Технические подписи для кабинета сотрудника */
export const BANK_STATUS_LABEL: Record<ApplicationStatus, string> = {
	DRAFT: 'Черновик',
	SUBMITTED: 'Подана',
	DATA_CHECK: 'Проверка данных',
	INFO_REQUESTED: 'Нужна доп. информация',
	SCORING: 'Скоринг',
	ON_DECISION: 'На решении',
	APPROVED: 'Одобрено',
	COUNTER_OFFER: 'Иные условия',
	REJECTED: 'Отказ',
	SIGNING: 'Подписание',
	ISSUED: 'Выдано',
	CLOSED: 'Закрыта',
};

/** Куда вести клиента по заявке: черновик — продолжить в мастере, остальное — в карточку */
export const applicationHref = (application: Pick<ApplicationShort, 'id' | 'status'>): string => (
	application.status === 'DRAFT' ? paths.applicationEdit(application.id) : paths.application(application.id)
);
