import { DASHBOARD_DEFINITIONS, type DashboardWidgetKind } from '@entities/dashboard';
import { Button, Modal, SearchField } from '@shared/ui';
import { useState } from 'react';

import styles from './Customization.module.scss';

interface WidgetCatalogProps { count: number; onAdd: (kind: DashboardWidgetKind) => void; onClose: () => void }
const kinds = Object.keys(DASHBOARD_DEFINITIONS) as DashboardWidgetKind[];
export const WidgetCatalog = ({ count, onAdd, onClose }: WidgetCatalogProps) => {
	const [search, setSearch] = useState('');
	const visible = kinds.filter((kind) => {
		const definition = DASHBOARD_DEFINITIONS[kind];
		return `${definition.title} ${definition.description}`.toLocaleLowerCase('ru-RU').includes(search.trim().toLocaleLowerCase('ru-RU'));
	});
	return <Modal title="Что добавить на главную?" size="large" onClose={onClose}>
		<div className={styles.catalogBody}>
			<p className={styles.hint}>Выберите виджет, затем настройте его содержание. Можно добавить несколько кредитов или заявок.</p>
			<SearchField label="Поиск виджетов" value={search} onChange={setSearch} placeholder="Кредит, заявка или календарь" />
			{count >= 30 && <p role="status">На главной может быть не более 30 виджетов.</p>}
			<div className={styles.catalog}>{visible.map((kind) => {
				const definition = DASHBOARD_DEFINITIONS[kind];
				return <article key={kind} className={styles.catalogCard}>
					<div className={styles.preview} data-kind={kind} aria-hidden="true"><span /><span /><span /></div>
					<h3>{definition.title}</h3><p>{definition.description}</p>
					<small>{definition.sizes.map(([w, h]) => `${w} × ${h}`).join(' · ')}</small>
					<Button variant="secondary" disabled={count >= 30} onClick={() => onAdd(kind)}>Добавить</Button>
				</article>;
			})}</div>
			{visible.length === 0 && <p>Ничего не найдено. Попробуйте другое название.</p>}
		</div>
	</Modal>;
};
