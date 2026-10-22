import { useLayoutEffect, useRef } from 'react';

import { Field, type FieldProps } from '../Field';
import { caretAfterDigits, digitsBefore, groupRubles } from './format';

export interface MoneyFieldProps extends Omit<FieldProps, 'value' | 'onChange' | 'type' | 'inputMode' | 'onKeyDown'> {
	/** Целые рубли без разделителей, пустая строка — поле не заполнено. */
	value: string;
	onValueChange: (value: string) => void;
}

export const MoneyField = ({ value, onValueChange, ref, ...props }: MoneyFieldProps) => {
	const inputRef = useRef<HTMLInputElement>(null);
	const caret = useRef<number | null>(null);
	useLayoutEffect(() => {
		if (caret.current === null) return;
		const position = caretAfterDigits(groupRubles(value), caret.current);
		inputRef.current?.setSelectionRange(position, position);
		caret.current = null;
	}, [value]);

	return (
		<Field {...props} type="text" inputMode="numeric" value={groupRubles(value)}
			ref={(node) => {
				inputRef.current = node;
				if (typeof ref === 'function') ref(node);
				else if (ref) ref.current = node;
			}}
			onChange={(event) => {
				const text = event.target.value;
				const digits = text.replace(/\s/gu, '');
				if (!/^\d*$/u.test(digits)) return;
				caret.current = digitsBefore(text, event.target.selectionStart ?? text.length);
				onValueChange(digits);
			}}
			onKeyDown={(event) => {
				const node = event.currentTarget;
				const start = node.selectionStart;
				if (start === null || start !== node.selectionEnd) return;
				const back = event.key === 'Backspace' && node.value[start - 1] === ' ';
				const next = event.key === 'Delete' && node.value[start] === ' ';
				if (!back && !next) return;
				event.preventDefault();
				const index = digitsBefore(node.value, start) - (back ? 1 : 0);
				caret.current = index;
				onValueChange(value.slice(0, index) + value.slice(index + 1));
			}} />
	);
};
