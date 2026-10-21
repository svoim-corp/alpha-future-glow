import { type ReactNode, type SVGProps } from 'react';

export type IconName =
	| 'home'
	| 'calendar'
	| 'applications'
	| 'loans'
	| 'catalog'
	| 'documents'
	| 'support'
	| 'settings'
	| 'logout'
	| 'search'
	| 'moon'
	| 'sun'
	| 'globe'
	| 'chevron'
	| 'arrow'
	| 'grid'
	| 'check'
	| 'building'
	| 'menu'
	| 'close'
	| 'bell';

/** Контуры иконок из макетов (docs/designs/shell.html), сетка 24×24 */
const SHAPES: Record<IconName, ReactNode> = {
	'home': <path d="m3 10 9-7 9 7v10a1 1 0 0 1-1 1h-5v-7H9v7H4a1 1 0 0 1-1-1z"/>,
	'calendar': <><rect x="3" y="5" width="18" height="16" rx="3"/><path d="M7 3v4m10-4v4M3 11h18m-13 4h2m4 0h2m-8 3h2"/></>,
	'applications': <><rect x="5" y="4" width="14" height="17" rx="2"/><path d="M9 4V2h6v2M9 10h6m-6 4h6m-6 4h3"/></>,
	'loans': <><rect x="3" y="5" width="18" height="15" rx="3"/><path d="M3 10h18m-14 5h4m6 0h1"/></>,
	'catalog': <><rect x="3" y="3" width="7" height="7" rx="2"/><rect x="14" y="3" width="7" height="7" rx="2"/><rect x="3" y="14" width="7" height="7" rx="2"/><rect x="14" y="14" width="7" height="7" rx="2"/></>,
	'documents': <><path d="M14 3H6a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"/><path d="M14 3v6h6M8 13h8m-8 4h5"/></>,
	'support': <><path d="M4 14v-3a8 8 0 0 1 16 0v3m-1 4v1a3 3 0 0 1-3 3h-3"/><rect x="2" y="11" width="5" height="8" rx="2"/><rect x="17" y="11" width="5" height="8" rx="2"/></>,
	'settings': <><path d="m9 3 1-1h4l1 3 3 1 3 1v4l-2 2v3l-2 3h-3l-2 2-3-2H6l-1-3-3-2v-4l3-1 1-3z"/><circle cx="12" cy="12" r="3"/></>,
	'logout': <path d="M9 4H5a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h4m7-13 5 5-5 5M8 12h13"/>,
	'search': <><circle cx="10.5" cy="10.5" r="6.5"/><path d="m16 16 5 5"/></>,
	'moon': <path d="M20.5 13A8.5 8.5 0 0 1 11 3.5 8.5 8.5 0 1 0 20.5 13Z"/>,
	'sun': <><circle cx="12" cy="12" r="4"/><path d="M12 2v2m0 16v2M2 12h2m16 0h2M5 5l1.5 1.5m11 11L19 19M5 19l1.5-1.5m11-11L19 5"/></>,
	'globe': <><circle cx="12" cy="12" r="9"/><ellipse cx="12" cy="12" rx="4" ry="9"/><path d="M3 12h18"/></>,
	'chevron': <path d="m8 10 4 4 4-4"/>,
	'arrow': <path d="M5 12h14m-5-5 5 5-5 5"/>,
	'grid': <><rect x="3" y="3" width="7" height="7" rx="1"/><rect x="14" y="3" width="7" height="7" rx="1"/><rect x="3" y="14" width="7" height="7" rx="1"/><rect x="14" y="14" width="7" height="7" rx="1"/></>,
	'check': <path d="m5 12 4 4L19 6"/>,
	'building': <><rect x="5" y="3" width="14" height="18" rx="2"/><path d="M9 7h1m4 0h1m-6 4h1m4 0h1m-6 4h1m4 0h1m-5 6v-3h2v3"/></>,
	'menu': <path d="M4 6h16M4 12h16M4 18h16"/>,
	'close': <path d="m6 6 12 12M6 18 18 6"/>,
	'bell': <path d="M6 8a6 6 0 0 1 12 0c0 7 3 9 3 9H3s3-2 3-9m7.3 13a2 2 0 0 1-3.4 0" />,
};

export interface IconProps extends Omit<SVGProps<SVGSVGElement>, 'name'> {
	name: IconName;
	/** Размер в пикселях (по умолчанию 22, как в макетах) */
	size?: number;
}

/** Линейная иконка: декоративная, скрыта от скринридеров — смысл несёт соседний текст или aria-label */
export const Icon = ({ name, size = 22, ...rest }: IconProps) => (
	<svg
		width={size}
		height={size}
		viewBox="0 0 24 24"
		fill="none"
		stroke="currentColor"
		strokeWidth={1.65}
		strokeLinecap="round"
		strokeLinejoin="round"
		aria-hidden="true"
		focusable="false"
		style={{ flexShrink: 0 }}
		{...rest}
	>
		{SHAPES[name]}
	</svg>
);
