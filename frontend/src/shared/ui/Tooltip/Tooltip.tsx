import { type ReactNode, useEffect, useId, useRef, useState } from 'react';
import { createPortal } from 'react-dom';

import { Icon } from '../Icon';
import styles from './Tooltip.module.scss';

export interface TooltipProps {
	label: string;
	content: ReactNode;
	children?: ReactNode;
}

/** Подсказка доступна при наведении, фокусе и касании; Escape закрывает её. */
export const Tooltip = ({ label, content, children }: TooltipProps) => {
	const id = useId();
	const trigger = useRef<HTMLButtonElement>(null);
	const bubble = useRef<HTMLDivElement>(null);
	const timer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
	const [position, setPosition] = useState<{
		left: number; top: number; bottom: number; above: boolean;
	} | null>(null);
	const cancelClose = () => clearTimeout(timer.current);
	const close = () => { cancelClose(); setPosition(null); };
	const show = () => {
		cancelClose();
		const rect = trigger.current?.getBoundingClientRect();
		if (!rect) return;
		const width = Math.min(300, window.innerWidth - 32);
		setPosition({
			left: Math.max(16, Math.min(rect.left + rect.width / 2 - width / 2, window.innerWidth - width - 16)),
			top: rect.bottom + 10,
			bottom: window.innerHeight - rect.top + 10,
			above: window.innerHeight - rect.bottom < 150 && rect.top > 150,
		});
	};
	const scheduleClose = () => {
		cancelClose();
		timer.current = setTimeout(() => {
			if (document.activeElement !== trigger.current) setPosition(null);
		}, 120);
	};

	useEffect(() => () => clearTimeout(timer.current), []);
	useEffect(() => {
		if (!position) return;
		const dismiss = () => setPosition(null);
		const escape = (event: KeyboardEvent) => { if (event.key === 'Escape') dismiss(); };
		const outside = (event: PointerEvent) => {
			if (event.target instanceof Node
				&& !trigger.current?.contains(event.target) && !bubble.current?.contains(event.target)) dismiss();
		};
		window.addEventListener('keydown', escape);
		window.addEventListener('resize', dismiss);
		window.addEventListener('scroll', dismiss, true);
		document.addEventListener('pointerdown', outside);

		return () => {
			window.removeEventListener('keydown', escape);
			window.removeEventListener('resize', dismiss);
			window.removeEventListener('scroll', dismiss, true);
			document.removeEventListener('pointerdown', outside);
		};
	}, [position]);

	return (
		<>
			<button ref={trigger} type="button" className={styles.trigger} aria-label={label}
				aria-describedby={position ? id : undefined}
				onMouseEnter={show} onMouseLeave={scheduleClose} onFocus={show} onBlur={close}
				onClick={show}>
				{children ?? <Icon name="info" size={18} />}
			</button>
			{position && createPortal(
				<div ref={bubble} role="tooltip" id={id} className={styles.bubble}
					style={{ left: position.left, top: position.above ? undefined : position.top,
						bottom: position.above ? position.bottom : undefined }}
					onMouseEnter={cancelClose} onMouseLeave={scheduleClose}>
					{content}
				</div>, document.body,
			)}
		</>
	);
};
