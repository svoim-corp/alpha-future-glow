import { type KeyboardEvent, type PointerEvent, useRef } from 'react';

import { type GridItem } from '../../lib/grid';
import styles from './TileGrid.module.scss';

const sides = ['top', 'right', 'bottom', 'left'] as const;
type Side = typeof sides[number];
interface TileResizeHandlesProps {
	item: GridItem;
	title: string;
	disabled: boolean;
	sizes: readonly (readonly [number, number])[];
	columnStep: () => number;
	onPreview: (item: GridItem | null) => void;
	onCommit: (item: GridItem) => void;
}
const labels: Record<Side, string> = { top: 'Верхняя', right: 'Правая', bottom: 'Нижняя', left: 'Левая' };

/** Противоположная сторона закреплена, размеры ограничены разрешёнными вариантами. */
const candidates = (item: GridItem, side: Side, sizes: TileResizeHandlesProps['sizes']) => {
	const horizontal = side === 'left' || side === 'right';
	return sizes.filter(([w, h]) => horizontal ? h === item.h : w === item.w).map(([w, h]) => ({
		...item, w, h,
		x: side === 'left' ? item.x + item.w - w : item.x,
		y: side === 'top' ? item.y + item.h - h : item.y,
	})).filter((rect) => rect.x >= 0 && rect.x + rect.w <= 12 && rect.y >= 0 && rect.y <= 200);
};

export const TileResizeHandles = ({ item, title, disabled, sizes, columnStep, onPreview, onCommit }: TileResizeHandlesProps) => {
	const gesture = useRef<{ pointerId: number; x: number; y: number; step: number; next: GridItem } | null>(null);
	const cancel = () => { gesture.current = null; onPreview(null); };
	const keyboard = (event: KeyboardEvent<HTMLButtonElement>, side: Side) => {
		if (event.key === 'Escape') { event.preventDefault(); cancel(); return; }
		const horizontal = side === 'left' || side === 'right';
		const delta = event.key === (horizontal ? 'ArrowLeft' : 'ArrowUp') ? -1
			: event.key === (horizontal ? 'ArrowRight' : 'ArrowDown') ? 1 : 0;
		if (!delta || disabled || gesture.current) return;
		event.preventDefault();
		const sign = side === 'left' || side === 'top' ? -1 : 1;
		const currentSize = horizontal ? item.w : item.h;
		const next = candidates(item, side, sizes)
			.filter((rect) => ((horizontal ? rect.w : rect.h) - currentSize) * delta * sign > 0)
			.sort((a, b) => Math.abs((horizontal ? a.w : a.h) - currentSize) - Math.abs((horizontal ? b.w : b.h) - currentSize))[0];
		if (next) onCommit(next);
	};
	const move = (event: PointerEvent<HTMLButtonElement>, side: Side) => {
		const active = gesture.current;
		if (!active || active.pointerId !== event.pointerId || disabled) return;
		const horizontal = side === 'left' || side === 'right';
		const sign = side === 'left' || side === 'top' ? -1 : 1;
		const delta = horizontal ? (event.clientX - active.x) / active.step : (event.clientY - active.y) / 80;
		const target = (horizontal ? item.w : item.h) + delta * sign;
		const options = candidates(item, side, sizes);
		const next = options.sort((a, b) => Math.abs((horizontal ? a.w : a.h) - target) - Math.abs((horizontal ? b.w : b.h) - target))[0];
		if (next) { active.next = next; onPreview(next); }
	};
	return <>{sides.map((side) => <button key={side} type="button" className={styles.resizeHandle} data-side={side}
		disabled={disabled || candidates(item, side, sizes).every((rect) => rect.w === item.w && rect.h === item.h)}
		aria-label={`${labels[side]} ручка размера: ${title}. Размер ${item.w} на ${item.h} клеток`}
		title={`${labels[side]} сторона — потяните или используйте стрелки`}
		onKeyDown={(event) => keyboard(event, side)}
		onPointerDown={(event) => {
			if (disabled || event.button !== 0) return;
			event.preventDefault(); event.stopPropagation();
			event.currentTarget.focus(); event.currentTarget.setPointerCapture(event.pointerId);
			gesture.current = { pointerId: event.pointerId, x: event.clientX, y: event.clientY, step: columnStep(), next: item };
		}}
		onPointerMove={(event) => move(event, side)}
		onPointerUp={(event) => {
			const active = gesture.current;
			if (!active || active.pointerId !== event.pointerId) return;
			gesture.current = null;
			if (!disabled && (active.next.w !== item.w || active.next.h !== item.h)) onCommit(active.next);
			onPreview(null);
			event.currentTarget.releasePointerCapture(event.pointerId);
		}}
		onPointerCancel={cancel} onLostPointerCapture={cancel}><span /></button>)}</>;
};
