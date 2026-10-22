import { type CSSProperties, type DragEvent, type ReactNode, useId, useRef, useState, useSyncExternalStore } from 'react';

import { orderedGrid, projectGrid, type GridItem } from '../../lib/grid';
import { Icon } from '../Icon';
import styles from './TileGrid.module.scss';

const subscribe = (notify: () => void) => {
	const queries = [window.matchMedia('(max-width:1250px)'), window.matchMedia('(max-width:600px)')];
	queries.forEach((query) => query.addEventListener('change', notify));
	return () => queries.forEach((query) => query.removeEventListener('change', notify));
};
const breakpoint = (): 1 | 6 | 12 => window.matchMedia('(max-width:600px)').matches ? 1 : window.matchMedia('(max-width:1250px)').matches ? 6 : 12;
interface TileGridProps<T extends GridItem> {
	items: readonly T[]; editing: boolean; disabled?: boolean;
	label: (item: T) => string; render: (item: T) => ReactNode;
	onMove: (id: string, x: number, y: number) => void;
	onReorder: (id: string, index: number) => void;
	onConfigure: (id: string) => void; onRemove: (id: string) => void;
}
export const TileGrid = <T extends GridItem>({ items, editing, disabled = false, label, render, onMove, onReorder, onConfigure, onRemove }: TileGridProps<T>) => {
	const columns = useSyncExternalStore(subscribe, breakpoint, () => 12 as const);
	const instructionsId = useId();
	const board = useRef<HTMLDivElement>(null);
	const dragging = useRef<string | null>(null);
	const [moving, setMoving] = useState<string | null>(null);
	const [preview, setPreview] = useState<GridItem | null>(null);
	const sequence = orderedGrid(items);
	const projection = columns === 1 ? sequence : projectGrid(sequence, columns);
	const clearDrag = () => { dragging.current = null; setMoving(null); setPreview(null); };
	const position = (event: DragEvent<HTMLDivElement>) => {
		const item = items.find((entry) => entry.id === dragging.current);
		const rect = board.current?.getBoundingClientRect();
		if (!item || !rect) return null;
		return { ...item, x: Math.max(0, Math.min(12 - item.w, Math.floor((event.clientX - rect.left) / ((rect.width + 16) / 12)))),
			y: Math.max(0, Math.min(200, Math.floor((event.clientY - rect.top) / 80))) };
	};
	return <>
		{editing && <p id={instructionsId} className={styles.instructions}>Перетащите виджет за ручку. На ручке доступны клавиши ↑ и ↓. Размер и порядок можно изменить в настройках виджета.</p>}
		<div ref={board} className={styles.grid} data-columns={columns} data-editing={editing} role="list" aria-label="Виджеты главной"
			onDragOver={(event) => { if (editing && !disabled && dragging.current) { event.preventDefault(); if (columns === 12) setPreview(position(event)); } }}
			onDrop={(event) => {
				const id = dragging.current;
				if (!editing || disabled || !id) return;
				event.preventDefault();
				if (columns === 12) { const next = position(event); if (next) onMove(id, next.x, next.y); }
				else {
					const target = event.target instanceof Element ? event.target.closest<HTMLElement>('[data-grid-id]')?.dataset.gridId : undefined;
					const index = sequence.filter((item) => item.id !== id).findIndex((item) => item.id === target);
					onReorder(id, index < 0 ? sequence.length - 1 : index);
				}
				clearDrag();
			}}>
			{projection.map((item) => {
				const original = items.find((entry) => entry.id === item.id) ?? item;
				const title = label(original); const index = sequence.findIndex((entry) => entry.id === item.id);
				const placement: CSSProperties = { gridColumn: `${item.x + 1} / span ${item.w}`, gridRow: `${item.y + 1} / span ${item.h}` };
				return <div key={item.id} role="listitem" data-grid-id={item.id} data-moving={moving === item.id} data-editing={editing} className={styles.tile} style={placement}>
					{editing && <div className={styles.tools} role="group" aria-label={`Настройка: ${title}`}>
						<button type="button" draggable={!disabled && columns !== 1} disabled={disabled}
							aria-label={`Перетащить: ${title}`} aria-describedby={instructionsId}
							onKeyDown={(event) => {
								if (event.key !== 'ArrowUp' && event.key !== 'ArrowDown') return;
								event.preventDefault();
								onReorder(item.id, Math.max(0, Math.min(sequence.length - 1, index + (event.key === 'ArrowUp' ? -1 : 1))));
							}}
							onDragStart={(event) => { dragging.current = item.id; setMoving(item.id); event.dataTransfer.effectAllowed = 'move'; event.dataTransfer.setData('text/plain', item.id);
								const tile = event.currentTarget.closest<HTMLElement>('[data-grid-id]'); if (tile) event.dataTransfer.setDragImage(tile, 24, 12); }}
							onDragEnd={clearDrag}>⠿</button>
						<button type="button" disabled={disabled} aria-label={`Настроить: ${title}`} onClick={() => onConfigure(item.id)}><Icon name="settings" size={16} /></button>
						<button type="button" disabled={disabled} aria-label={`Убрать: ${title}`} onClick={() => onRemove(item.id)}><Icon name="close" size={16} /></button>
					</div>}
					{render(original)}
				</div>;
			})}
			{editing && preview && columns === 12 && <div className={styles.preview} aria-hidden="true" style={{ gridColumn: `${preview.x + 1} / span ${preview.w}`, gridRow: `${preview.y + 1} / span ${preview.h}` }} />}
		</div>
	</>;
};
