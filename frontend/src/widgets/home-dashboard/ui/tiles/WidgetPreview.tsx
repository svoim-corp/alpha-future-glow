import { type CSSProperties, type ReactNode, useLayoutEffect, useRef, useState } from 'react';

import styles from '../Dashboard.module.scss';

interface WidgetPreviewProps {
	children: ReactNode[];
	rowHeight: number;
	minColumnWidth?: number;
	maxColumns?: number;
}

/** Количество записей зависит от свободного места, скрытые записи не попадают в Tab-порядок. */
export const WidgetPreview = ({ children, rowHeight, minColumnWidth = 240, maxColumns = 1 }: WidgetPreviewProps) => {
	const ref = useRef<HTMLDivElement>(null);
	const [layout, setLayout] = useState({ count: 1, columns: 1, natural: false });
	useLayoutEffect(() => {
		const element = ref.current;
		if (!element) return;
		const update = () => {
			const natural = window.matchMedia('(max-width:600px)').matches;
			const columns = Math.max(1, Math.min(maxColumns, Math.floor((element.clientWidth + 8) / (minColumnWidth + 8))));
			const count = natural ? Math.min(children.length, 3) : columns * Math.max(0, Math.floor((element.clientHeight + 8) / (rowHeight + 8)));
			setLayout((previous) => previous.count === count && previous.columns === columns && previous.natural === natural
				? previous : { count, columns, natural });
		};
		update();
		const observer = new ResizeObserver(update);
		observer.observe(element);
		window.addEventListener('resize', update);
		return () => { observer.disconnect(); window.removeEventListener('resize', update); };
	}, [children.length, maxColumns, minColumnWidth, rowHeight]);
	const style: CSSProperties = {
		gridTemplateColumns: `repeat(${layout.columns}, minmax(0, 1fr))`,
		gridAutoRows: layout.natural ? undefined : `${rowHeight}px`,
	};
	return <div ref={ref} className={styles.previewList} style={style}>
		{children.slice(0, layout.count)}
	</div>;
};
