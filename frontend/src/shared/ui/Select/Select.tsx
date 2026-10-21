import clsx from 'clsx';
import { type KeyboardEvent, useEffect, useId, useRef, useState } from 'react';

import { Icon } from '../Icon';
import styles from './Select.module.scss';

export interface SelectOption {
	value: string;
	label: string;
}

export type SelectVariant = 'field' | 'filter' | 'inline';

export interface SelectProps {
	options: SelectOption[];
	value: string;
	onChange: (value: string) => void;
	id?: string;
	name?: string;
	variant?: SelectVariant;
	disabled?: boolean;
	invalid?: boolean;
	className?: string;
	'aria-label'?: string;
	'aria-labelledby'?: string;
	'aria-describedby'?: string;
}

/** Высота списка, при которой он ещё открывается вниз, иначе — вверх */
const PANEL_SPACE = 260;

/**
 * Выпадающий список вместо нативного select: тот же контракт (value/onChange),
 * но панель оформлена токенами UI-кита и одинакова в светлой и тёмной теме.
 * Клавиатура: стрелки, Home/End, Enter/Space, Escape, первая буква варианта.
 */
export const Select = ({
	options,
	value,
	onChange,
	id,
	name,
	variant = 'field',
	disabled,
	invalid,
	className,
	'aria-label': ariaLabel,
	'aria-labelledby': ariaLabelledBy,
	'aria-describedby': ariaDescribedBy,
}: SelectProps) => {
	const generated = useId();
	const listId = `${generated}-list`;
	const rootRef = useRef<HTMLDivElement>(null);
	const listRef = useRef<HTMLUListElement>(null);
	const [open, setOpen] = useState(false);
	const [openUp, setOpenUp] = useState(false);
	const selectedIndex = Math.max(0, options.findIndex((option) => option.value === value));
	const [active, setActive] = useState(selectedIndex);
	const selected = options.find((option) => option.value === value);

	useEffect(() => {
		if (!open) return undefined;
		const close = (event: MouseEvent) => {
			if (!rootRef.current?.contains(event.target as Node)) setOpen(false);
		};
		document.addEventListener('mousedown', close);
		return () => document.removeEventListener('mousedown', close);
	}, [open]);

	useEffect(() => {
		if (!open) return;
		const node = listRef.current?.children[active];
		if (node && 'scrollIntoView' in node) (node as HTMLElement).scrollIntoView({ block: 'nearest' });
	}, [open, active]);

	const show = () => {
		if (disabled) return;
		const rect = rootRef.current?.getBoundingClientRect();
		setOpenUp(rect !== undefined && window.innerHeight - rect.bottom < PANEL_SPACE && rect.top > PANEL_SPACE);
		setActive(selectedIndex);
		setOpen(true);
	};

	const choose = (index: number) => {
		const option = options[index];
		setOpen(false);
		if (option && option.value !== value) onChange(option.value);
	};

	const move = (index: number) => setActive(Math.min(options.length - 1, Math.max(0, index)));

	const onKeyDown = (event: KeyboardEvent<HTMLButtonElement>) => {
		if (disabled) return;
		const { key } = event;
		if (key === 'Tab') {
			setOpen(false);
			return;
		}
		if (key === 'Escape') {
			if (open) {
				event.preventDefault();
				event.stopPropagation();
				setOpen(false);
			}
			return;
		}
		if (key === 'ArrowDown' || key === 'ArrowUp') {
			event.preventDefault();
			if (!open) show();
			else move(active + (key === 'ArrowDown' ? 1 : -1));
			return;
		}
		if (key === 'Home' || key === 'End') {
			event.preventDefault();
			if (!open) show();
			move(key === 'Home' ? 0 : options.length - 1);
			return;
		}
		if (key === 'Enter' || key === ' ') {
			event.preventDefault();
			if (open) choose(active);
			else show();
			return;
		}
		if (key.length === 1 && !event.ctrlKey && !event.metaKey && !event.altKey) {
			const letter = key.toLowerCase();
			const from = open ? active + 1 : selectedIndex + 1;
			const ordered = [...options.slice(from), ...options.slice(0, from)];
			const found = ordered.find((option) => option.label.toLowerCase().startsWith(letter));
			if (!found) return;
			const index = options.indexOf(found);
			if (open) setActive(index);
			else if (found.value !== value) onChange(found.value);
		}
	};

	return (
		<div ref={rootRef} className={clsx(styles.root, styles[variant], className)}>
			<button
				type="button"
				id={id}
				role="combobox"
				className={styles.trigger}
				aria-haspopup="listbox"
				aria-expanded={open}
				aria-controls={open ? listId : undefined}
				aria-activedescendant={open ? `${generated}-opt-${active}` : undefined}
				aria-label={ariaLabel}
				aria-labelledby={ariaLabelledBy}
				aria-describedby={ariaDescribedBy}
				aria-invalid={invalid ? true : undefined}
				disabled={disabled}
				onClick={() => (open ? setOpen(false) : show())}
				onKeyDown={onKeyDown}
			>
				<span className={styles.value}>{selected?.label ?? ''}</span>
				<Icon name="chevron" size={18} className={styles.chevron} />
			</button>
			{name && <input type="hidden" name={name} value={value} />}
			{open && (
				<ul
					ref={listRef}
					id={listId}
					role="listbox"
					tabIndex={-1}
					aria-label={ariaLabel}
					aria-labelledby={ariaLabelledBy}
					className={clsx(styles.panel, openUp && styles.up)}
					onMouseDown={(event) => event.preventDefault()}
					onPointerUp={(event) => {
						const item = (event.target as HTMLElement).closest<HTMLElement>('[role="option"]');
						if (item) choose(Number(item.dataset.index));
					}}
				>
					{options.map((option, index) => (
						<li
							key={option.value}
							id={`${generated}-opt-${index}`}
							role="option"
							aria-selected={option.value === value}
							data-index={index}
							className={clsx(
								styles.option,
								index === active && styles.active,
								option.value === value && styles.selected,
							)}
							onMouseEnter={() => setActive(index)}
						>
							<span>{option.label}</span>
							{option.value === value && <Icon name="check" size={16} />}
						</li>
					))}
				</ul>
			)}
		</div>
	);
};
