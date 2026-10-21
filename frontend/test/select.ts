import { screen } from '@testing-library/react';
import { type UserEvent } from '@testing-library/user-event';

/** Выбор варианта в кастомном выпадающем списке: открыть по имени поля и кликнуть вариант по подписи */
export const chooseOption = async (user: UserEvent, fieldName: string, optionName: string) => {
	await user.click(screen.getByRole('combobox', { name: fieldName }));
	await user.click(await screen.findByRole('option', { name: optionName }));
};
