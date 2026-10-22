/** Приветствие по времени суток */
export const greeting = (hour: number) => {
	if (hour < 5) {
		return 'Доброй ночи';
	}

	if (hour < 12) {
		return 'Доброе утро';
	}

	return hour < 18 ? 'Добрый день' : 'Добрый вечер';
};

/** Имя для приветствия: в ФИО «Фамилия Имя Отчество» имя — второе слово */
export const firstName = (fullName: string) => fullName.split(' ').filter(Boolean)[1] ?? fullName;
