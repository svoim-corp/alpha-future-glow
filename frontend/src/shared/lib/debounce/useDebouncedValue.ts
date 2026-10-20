import { useEffect, useState } from 'react';

/** Значение, которое обновляется после паузы в изменениях: не дёргаем API на каждый символ */
export const useDebouncedValue = <T>(value: T, delayMs = 400): T => {
	const [debounced, setDebounced] = useState(value);

	useEffect(() => {
		const timer = setTimeout(() => setDebounced(value), delayMs);

		return () => clearTimeout(timer);
	}, [value, delayMs]);

	return debounced;
};
