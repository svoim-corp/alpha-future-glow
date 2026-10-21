import { useLayoutEffect, useRef, useState } from 'react';

/** Сохраняет место под результатом при загрузке; при смене ширины измеряет высоту заново. */
export const useStableHeight = () => {
	const contentRef = useRef<HTMLDivElement>(null);
	const [size, setSize] = useState({ width: 0, height: 0 });

	useLayoutEffect(() => {
		const content = contentRef.current;
		if (!content) return undefined;

		const measure = () => {
			const bounds = content.getBoundingClientRect();
			const width = Math.round(bounds.width);
			const height = Math.ceil(bounds.height);
			if (width === 0) return;
			setSize((previous) => {
				const nextHeight = previous.width === width ? Math.max(previous.height, height) : height;
				return previous.width === width && previous.height === nextHeight
					? previous : { width, height: nextHeight };
			});
		};

		measure();
		if (typeof ResizeObserver === 'undefined') return undefined;
		const observer = new ResizeObserver(measure);
		observer.observe(content);
		return () => observer.disconnect();
	});

	return { contentRef, minHeight: size.height };
};
