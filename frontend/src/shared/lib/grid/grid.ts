export interface GridRect { x: number; y: number; w: number; h: number }
export interface GridItem extends GridRect { id: string }

export const intersects = (a: GridRect, b: GridRect) =>
	a.x < b.x + b.w && a.x + a.w > b.x && a.y < b.y + b.h && a.y + a.h > b.y;

export const orderedGrid = <T extends GridItem>(items: readonly T[]): T[] =>
	[...items].sort((a, b) => a.y - b.y || a.x - b.x || a.id.localeCompare(b.id));

export const vacantPosition = (size: Pick<GridRect, 'w' | 'h'>, occupied: readonly GridRect[], columns = 12): GridRect => {
	if (!Number.isInteger(size.w) || !Number.isInteger(size.h) || size.w > columns || size.w < 1 || size.h < 1) throw new Error('Недопустимый размер элемента сетки');
	for (let y = 0; ; y += 1) {
		for (let x = 0; x <= columns - size.w; x += 1) {
			const rect = { ...size, x, y };
			if (!occupied.some((item) => intersects(rect, item))) return rect;
		}
	}
};

/** Закреплённый элемент сохраняет позицию; пересекающиеся детерминированно переупаковываются. */
export const arrangeGrid = <T extends GridItem>(items: readonly T[], anchor?: T): T[] => {
	const placed: T[] = anchor ? [{ ...anchor }] : [];
	for (const item of orderedGrid(items).filter((entry) => entry.id !== anchor?.id)) {
		placed.push(placed.some((entry) => intersects(item, entry))
			? { ...item, ...vacantPosition(item, placed) } : { ...item });
	}
	return orderedGrid(placed);
};

/** Последовательная упаковка сохраняет одинаковый визуальный и клавиатурный порядок. */
const flowGrid = <T extends GridItem>(items: readonly T[], columns: number): T[] => {
	let x = 0; let y = 0; let rowHeight = 0;
	return items.map((item) => {
		if (x + item.w > columns) { x = 0; y += rowHeight; rowHeight = 0; }
		const next = { ...item, x, y };
		x += item.w; rowHeight = Math.max(rowHeight, item.h);
		return next;
	});
};

/** Меньший экран меняет только проекцию, никогда не исходную конфигурацию. */
export const projectGrid = <T extends GridItem>(items: readonly T[], columns: 6 | 12): T[] => {
	if (columns === 12) return orderedGrid(items);
	return flowGrid(orderedGrid(items).map((item) => ({ ...item, w: item.w >= 6 ? 6 : 3 })), columns);
};

export const reorderGrid = <T extends GridItem>(items: readonly T[], id: string, target: number): T[] => {
	const sequence = orderedGrid(items);
	const index = sequence.findIndex((item) => item.id === id);
	const current = sequence[index];
	if (!current) return sequence;
	sequence.splice(index, 1);
	sequence.splice(Math.max(0, Math.min(sequence.length, target)), 0, current);
	return flowGrid(sequence, 12);
};
