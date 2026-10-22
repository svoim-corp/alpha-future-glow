import clsx from 'clsx';
import { type ReactNode, useEffect, useId, useRef } from 'react';
import { createPortal } from 'react-dom';

import { Icon } from '../Icon';
import styles from './Modal.module.scss';

export interface ModalProps {
	title: string;
	onClose: () => void;
	children: ReactNode;
	footer?: ReactNode;
	wide?: boolean;
	size?: 'small' | 'default' | 'wide' | 'large';
	layer?: 'default' | 'overlay';
}

const FOCUSABLE = 'a[href], button:not(:disabled), input:not(:disabled), select:not(:disabled), textarea:not(:disabled)';

/**
 * Диалог: фокус внутри, Escape закрывает, после закрытия фокус возвращается на прежний элемент.
 * Рендерится в portal, чтобы наследовать тему с body.
 */
export const Modal = ({ title, onClose, children, footer, wide = false, size, layer = 'default' }: ModalProps) => {
	const titleId = useId();
	const dialogRef = useRef<HTMLDivElement>(null);
	const variant = size ?? (wide ? 'wide' : 'default');

	useEffect(() => {
		const previous = document.activeElement instanceof HTMLElement ? document.activeElement : null;
		const underlying = Array.from(document.querySelectorAll<HTMLElement>('[role="dialog"][aria-modal="true"]'))
			.filter((dialog) => dialog !== dialogRef.current)
			.map((dialog) => ({ dialog, hidden: dialog.getAttribute('aria-hidden'), inert: dialog.inert }));
		dialogRef.current?.focus();
		for (const { dialog } of underlying) {
			dialog.inert = true;
			dialog.setAttribute('aria-hidden', 'true');
		}

		return () => {
			for (const { dialog, hidden, inert } of underlying) {
				dialog.inert = inert;
				if (hidden === null) dialog.removeAttribute('aria-hidden');
				else dialog.setAttribute('aria-hidden', hidden);
			}
			previous?.focus();
		};
	}, []);

	const onCloseRef = useRef(onClose);

	useEffect(() => {
		onCloseRef.current = onClose;
	});

	useEffect(() => {
		const dialog = dialogRef.current;

		if (!dialog) {
			return undefined;
		}

		const handleKeyDown = (event: KeyboardEvent) => {
			if (event.key === 'Escape') {
				// Открытое меню сначала обрабатывает Escape своим обработчиком React.
				if (event.target instanceof Element
					&& event.target.closest('[role="combobox"][aria-expanded="true"]')) return;
				event.stopPropagation();
				onCloseRef.current();

				return;
			}

			if (event.key !== 'Tab') {
				return;
			}

			const items = Array.from(dialog.querySelectorAll<HTMLElement>(FOCUSABLE))
				.filter((item) => item.tabIndex >= 0 && item.getAttribute('type') !== 'hidden');
			const first = items[0];
			const last = items[items.length - 1];

			if (!first || !last) {
				return;
			}

			if (event.shiftKey && (document.activeElement === first || document.activeElement === dialog)) {
				event.preventDefault();
				last.focus();
			} else if (!event.shiftKey && document.activeElement === last) {
				event.preventDefault();
				first.focus();
			}
		};

		dialog.addEventListener('keydown', handleKeyDown);

		return () => dialog.removeEventListener('keydown', handleKeyDown);
	}, []);

	return createPortal(
		<div className={clsx(styles.overlay, layer === 'overlay' && styles.topLayer)}>
			<div
				ref={dialogRef}
				className={clsx(styles.dialog, variant === 'small' && styles.small,
					variant === 'wide' && styles.wide, variant === 'large' && styles.large)}
				role="dialog"
				aria-modal="true"
				aria-labelledby={titleId}
				tabIndex={-1}
			>
				<div className={styles.header}>
					<h2 id={titleId}>{title}</h2>
					<button type="button" className={styles.close} aria-label="Закрыть" onClick={onClose}>
						<Icon name="close" size={18} />
					</button>
				</div>
				<div className={styles.body}>{children}</div>
				{footer && <div className={styles.footer}>{footer}</div>}
			</div>
		</div>,
		document.body,
	);
};
