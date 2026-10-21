import clsx from 'clsx';
import { type ReactNode, useEffect, useId, useRef } from 'react';
import { createPortal } from 'react-dom';

import styles from './Modal.module.scss';

export interface ModalProps {
	title: string;
	onClose: () => void;
	children: ReactNode;
	footer?: ReactNode;
	wide?: boolean;
}

const FOCUSABLE = 'a[href], button:not(:disabled), input:not(:disabled), select:not(:disabled), textarea:not(:disabled)';

/**
 * Диалог: фокус внутри, Escape закрывает, после закрытия фокус возвращается на прежний элемент.
 * Рендерится в portal, чтобы наследовать тему с body.
 */
export const Modal = ({ title, onClose, children, footer, wide = false }: ModalProps) => {
	const titleId = useId();
	const dialogRef = useRef<HTMLDivElement>(null);

	useEffect(() => {
		const previous = document.activeElement instanceof HTMLElement ? document.activeElement : null;
		dialogRef.current?.focus();

		return () => previous?.focus();
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
				event.stopPropagation();
				onCloseRef.current();

				return;
			}

			if (event.key !== 'Tab') {
				return;
			}

			const items = Array.from(dialog.querySelectorAll<HTMLElement>(FOCUSABLE));
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
		<div className={styles.overlay}>
			<div
				ref={dialogRef}
				className={clsx(styles.dialog, wide && styles.wide)}
				role="dialog"
				aria-modal="true"
				aria-labelledby={titleId}
				tabIndex={-1}
			>
				<div className={styles.header}>
					<h2 id={titleId}>{title}</h2>
					<button type="button" className={styles.close} aria-label="Закрыть" onClick={onClose}>×</button>
				</div>
				<div>{children}</div>
				{footer && <div className={styles.footer}>{footer}</div>}
			</div>
		</div>,
		document.body,
	);
};
