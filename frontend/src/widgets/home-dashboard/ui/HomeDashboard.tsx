import { DASHBOARD_DEFINITIONS, type DashboardScope, type DashboardWidgetKind } from '@entities/dashboard';
import { useGetMeQuery } from '@entities/session';
import {
	addWidget, cancelEditing, moveWidget, removeWidget, reorderWidget,
	resetDraft, updateWidget, useDashboardEditor, WidgetCatalog, WidgetSettings,
} from '@features/customize-dashboard';
import { orderedGrid } from '@shared/lib/grid';
import { useAppDispatch } from '@shared/lib/store';
import { Alert, AsyncBoundary, Button, Icon, type IconName, Modal, TileGrid } from '@shared/ui';
import { useState } from 'react';
import { useBeforeUnload, useBlocker } from 'react-router';

import styles from './Dashboard.module.scss';
import { DashboardAttention } from './tiles/ApplicationWidgets';
import { WidgetContent } from './tiles/WidgetContent';

const icons: Record<DashboardWidgetKind, IconName> = {
	applications: 'applications', loan: 'loans', payment: 'calendar', debt: 'loans', quick: 'catalog',
	calendar: 'calendar', plans: 'calendar', application: 'documents',
	loans: 'loans', payments: 'calendar', offers: 'check', drafts: 'applications',
	planFunding: 'catalog', notifications: 'bell', tickets: 'support', funding: 'catalog',
};
const ScopedDashboard = ({ scope }: { scope: DashboardScope }) => {
	const dispatch = useAppDispatch();
	const editor = useDashboardEditor(scope);
	const [catalogOpen, setCatalogOpen] = useState(false);
	const [selectedId, setSelectedId] = useState<string | null>(null);
	const saving = editor.status === 'saving';
	const dirty = editor.hasChanges;
	const blocker = useBlocker(dirty);
	useBeforeUnload((event) => { if (dirty) { event.preventDefault(); event.returnValue = ''; } });
	const selected = editor.config.widgets.find((widget) => widget.id === selectedId);
	const selectedPosition = orderedGrid(editor.config.widgets).findIndex((widget) => widget.id === selectedId);
	const configure = (id: string) => {
		if (saving || editor.status !== 'ready') return;
		if (editor.beginEditing()) setSelectedId(id);
	};
	const openCatalog = () => { if (editor.beginEditing()) setCatalogOpen(true); };
	const add = (kind: DashboardWidgetKind) => {
		if (editor.config.widgets.length >= 30) return;
		const id = crypto.randomUUID();
		dispatch(addWidget({ kind, id })); setCatalogOpen(false); setSelectedId(id);
	};
	if (!editor.ready) return <div className={styles.root}>
		<DashboardAttention />
		{editor.error
			? <Alert tone="danger" title={editor.error} actions={<Button variant="secondary" onClick={editor.reload}>Повторить загрузку</Button>} />
			: <p role="status">Загрузка настройки главной…</p>}
	</div>;
	return <div className={styles.root}>
		<DashboardAttention />
		<div className={styles.toolbar}>
			<div><h2>Моё рабочее пространство</h2>{editor.editing && <p className={styles.muted}>Меняйте состав, расположение и размеры виджетов.</p>}</div>
			{editor.editing ? <div className={styles.toolbarActions}>
				<Button variant="ghost" disabled={saving} onClick={() => dispatch(resetDraft())}>Сбросить</Button>
				<Button variant="secondary" disabled={saving || editor.config.widgets.length >= 30} onClick={openCatalog}>Добавить виджет</Button>
				<Button variant="secondary" disabled={saving} onClick={() => dispatch(cancelEditing())}>Отмена</Button>
				<Button disabled={saving} onClick={() => void editor.save()}>{saving ? 'Сохраняем…' : 'Сохранить'}</Button>
			</div> : <div className={styles.toolbarActions}>
				{editor.canImport && <Button variant="ghost" disabled={editor.status !== 'ready'} onClick={editor.importLegacy}>Перенести настройку с устройства</Button>}
				<Button variant="secondary" disabled={editor.status !== 'ready'} onClick={editor.beginEditing}><Icon name="settings" size={18} />Настроить главную</Button>
			</div>}
		</div>
		{editor.error && <Alert tone="danger" title={editor.error} actions={<Button variant="secondary" disabled={saving} onClick={editor.reload}>
			{editor.editing ? 'Отменить изменения и загрузить сохранённую' : 'Повторить загрузку'}
		</Button>} />}
		<TileGrid items={editor.config.widgets} editing={editor.editing} disabled={saving}
			label={(widget) => DASHBOARD_DEFINITIONS[widget.type].title}
			resizeSizes={(widget) => DASHBOARD_DEFINITIONS[widget.type].sizes}
			onResize={(rect) => {
				const widget = editor.config.widgets.find((entry) => entry.id === rect.id);
				if (widget) dispatch(updateWidget({ ...widget, ...rect }));
			}}
			onMove={(id, x, y) => dispatch(moveWidget({ id, x, y }))}
			onReorder={(id, index) => dispatch(reorderWidget({ id, index }))}
			onConfigure={configure} onRemove={(id) => dispatch(removeWidget(id))}
			render={(widget) => <section className={styles.widget} data-kind={widget.type} data-height={widget.h} data-width={widget.w} aria-label={DASHBOARD_DEFINITIONS[widget.type].title}>
				<div className={styles.widgetHeading}><span className={styles.widgetIcon}><Icon name={icons[widget.type]} size={18} /></span><h3>{DASHBOARD_DEFINITIONS[widget.type].title}</h3></div>
				<div className={styles.body}><WidgetContent widget={widget} onConfigure={() => configure(widget.id)} /></div>
			</section>} />
		{editor.config.widgets.length === 0 && <p className={styles.empty}>Главная пока пуста. Добавьте виджеты, которые нужны вам.</p>}
		<div className={styles.footer}><p role="status" aria-live="polite">{editor.message || 'Сохраняйте настройку главной в своём профиле — она будет доступна на других устройствах.'}</p>
			<Button variant="ghost" disabled={saving || editor.status !== 'ready' || editor.config.widgets.length >= 30} onClick={openCatalog}>＋ Добавить виджет</Button>
		</div>
		{catalogOpen && editor.editing && <WidgetCatalog count={editor.config.widgets.length} onAdd={add} onClose={() => setCatalogOpen(false)} />}
		{selected && editor.editing && <WidgetSettings key={selected.id} widget={selected} position={selectedPosition} total={editor.config.widgets.length}
			onClose={() => setSelectedId(null)} onApply={(widget, position) => {
				dispatch(updateWidget(widget));
				if (position !== selectedPosition) dispatch(reorderWidget({ id: widget.id, index: position }));
				setSelectedId(null);
			}} />}
		{blocker.state === 'blocked' && <Modal title="Настройка ещё не сохранена" onClose={() => blocker.reset?.()}
			footer={<><Button variant="secondary" onClick={() => blocker.reset?.()}>Остаться</Button><Button disabled={saving} onClick={() => { dispatch(cancelEditing()); blocker.proceed?.(); }}>Уйти без сохранения</Button></>}>
			<p>На главной есть несохранённая настройка. Сохраните её перед переходом или отмените изменения.</p>
		</Modal>}
	</div>;
};

export const HomeDashboard = () => {
	const user = useGetMeQuery();
	return <AsyncBoundary isLoading={user.isLoading} isError={user.isError} onRetry={() => void user.refetch()} loadingLabel="Загрузка главной" errorText="Не удалось загрузить профиль.">
		{user.data?.company ? <ScopedDashboard key={`${user.data.id}:${user.data.company.id}`} scope={{ userId: user.data.id, companyId: user.data.company.id }} />
			: <p>Главная доступна после привязки компании к профилю.</p>}
	</AsyncBoundary>;
};
