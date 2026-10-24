import clsx from 'clsx';
import { type CSSProperties, type KeyboardEvent, useEffect, useId, useLayoutEffect, useRef, useState } from 'react';

import { createPortal } from 'react-dom';

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

/** Максимальная высота меню; при нехватке места список прокручивается. */
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
	const [position, setPosition] = useState<CSSProperties | null>(null);
	const selectedIndex = Math.max(0, options.findIndex((option) => option.value === value));
	const [active, setActive] = useState(selectedIndex);
	const selected = options.find((option) => option.value === value);

	useEffect(() => {
		if (!open) return undefined;
		const close = (event: PointerEvent) => {
			if (event.target instanceof Node && !rootRef.current?.contains(event.target)
				&& !listRef.current?.contains(event.target)) setOpen(false);
		};
		document.addEventListener('pointerdown', close);
		return () => document.removeEventListener('pointerdown', close);
	}, [open]);

	useLayoutEffect(() => {
		const panel = listRef.current;
		const root = rootRef.current;
		if (!open || !panel || !root) return;
		panel.showPopover();
		const update = () => {
			const rect = root.getBoundingClientRect();
			const viewport = window.visualViewport;
			const left = (viewport?.offsetLeft ?? 0) + 12;
			const top = (viewport?.offsetTop ?? 0) + 12;
			const right = left + (viewport?.width ?? window.innerWidth) - 24;
			const bottom = top + (viewport?.height ?? window.innerHeight) - 24;
			const width = Math.min(Math.max(rect.width, variant === 'field' ? 0 : 200), right - left);
			const below = Math.max(0, bottom - rect.bottom - 6);
			const above = Math.max(0, rect.top - top - 6);
			const desired = Math.min(PANEL_SPACE, panel.scrollHeight + 2);
			const up = below < desired && above > below;
			setPosition({
				left: Math.max(left, Math.min(rect.left, right - width)), width,
				top: up ? undefined : Math.max(top, rect.bottom + 6),
				bottom: up ? window.innerHeight - Math.min(bottom, rect.top - 6) : undefined,
				maxHeight: Math.min(PANEL_SPACE, up ? above : below),
			});
		};
		const scroll = (event: Event) => {
			if (event.target instanceof Node && panel.contains(event.target)) return;
			update();
		};
		update();
		const observer = new ResizeObserver(update);
		observer.observe(root);
		observer.observe(panel);
		window.addEventListener('resize', update);
		window.addEventListener('scroll', scroll, true);
		window.visualViewport?.addEventListener('resize', update);
		window.visualViewport?.addEventListener('scroll', update);
		return () => {
			observer.disconnect();
			window.removeEventListener('resize', update);
			window.removeEventListener('scroll', scroll, true);
			window.visualViewport?.removeEventListener('resize', update);
			window.visualViewport?.removeEventListener('scroll', update);
			panel.hidePopover();
		};
	}, [open, variant]);

	useEffect(() => {
		if (!open) return;
		const panel = listRef.current;
		const node = panel?.children[active];
		if (!panel || !(node instanceof HTMLElement)) return;
		// Прокручиваем только список, не страницу и не модальное окно.
		if (node.offsetTop < panel.scrollTop) panel.scrollTop = node.offsetTop;
		else if (node.offsetTop + node.offsetHeight > panel.scrollTop + panel.clientHeight) {
			panel.scrollTop = node.offsetTop + node.offsetHeight - panel.clientHeight;
		}
	}, [open, active]);

	const show = () => {
		if (disabled) return;
		setPosition(null);
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
			{open && createPortal(
				<ul
					popover="manual"
					style={{ ...position, visibility: position ? undefined : 'hidden' }}
					ref={listRef}
					id={listId}
					role="listbox"
					tabIndex={-1}
					aria-label={ariaLabel}
					aria-labelledby={ariaLabelledBy}
					className={styles.panel}
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
				</ul>,
				rootRef.current?.closest('[role="dialog"]') ?? document.body,
			)}
		</div>
	);
};
