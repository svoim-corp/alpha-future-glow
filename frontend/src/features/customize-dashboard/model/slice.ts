import {
	createDashboardWidget, createDefaultDashboard, dashboardConfigSchema,
	type DashboardConfig, type DashboardDocument, type DashboardScope,
	type DashboardWidget, type DashboardWidgetKind,
} from '@entities/dashboard';
import { createSlice, type PayloadAction } from '@reduxjs/toolkit';
import { baseApi } from '@shared/api';
import { arrangeGrid, reorderGrid, vacantPosition } from '@shared/lib/grid';

interface DashboardEditorState {
	scope: DashboardScope | null;
	/** Только исходный снимок редактирования; актуальный документ принадлежит RTK Query. */
	baseline: DashboardConfig | null;
	revision: string | null;
	draft: DashboardConfig | null;
	status: 'ready' | 'saving';
	error: string | null;
	message: string;
	requestId: string | null;
}
const initialState: DashboardEditorState = { scope: null, baseline: null, revision: null, draft: null, status: 'ready', error: null, message: '', requestId: null };
interface EditArgs { scope: DashboardScope; document: DashboardDocument | null; config?: DashboardConfig }

export const dashboardCustomizationSlice = createSlice({
	name: 'dashboardCustomization',
	initialState,
	reducers: {
		beginEditing: (state, { payload }: PayloadAction<EditArgs>) => {
			if (state.status === 'saving') return;
			if (state.draft && state.scope?.userId === payload.scope.userId && state.scope.companyId === payload.scope.companyId) return;
			state.scope = payload.scope;
			state.baseline = dashboardConfigSchema.parse(payload.document?.config ?? createDefaultDashboard());
			state.revision = payload.document?.revision ?? null;
			state.draft = dashboardConfigSchema.parse(payload.config ?? state.baseline);
			state.error = null; state.message = '';
		},
		cancelEditing: (state) => {
			if (state.status === 'saving') return;
			state.draft = null; state.baseline = null; state.revision = null;
			state.error = null; state.message = 'Настройка отменена.';
		},
		saveStarted: (state, { payload }: PayloadAction<string>) => {
			if (!state.draft || state.status === 'saving') return;
			state.status = 'saving'; state.error = null; state.requestId = payload;
		},
		saveCompleted: (state, { payload }: PayloadAction<string>) => {
			if (state.requestId !== payload) return;
			state.draft = null; state.baseline = null; state.revision = null;
			state.status = 'ready'; state.requestId = null;
			state.message = 'Главная сохранена в вашем профиле.';
		},
		saveFailed: (state, { payload }: PayloadAction<{ requestId: string; message: string }>) => {
			if (state.requestId !== payload.requestId) return;
			state.status = 'ready'; state.requestId = null; state.error = payload.message;
		},
		resetDraft: (state) => {
			if (state.draft && state.status === 'ready') { state.draft = createDefaultDashboard(); state.message = 'Исходная раскладка восстановлена в черновике.'; }
		},
		addWidget: (state, { payload }: PayloadAction<{ kind: DashboardWidgetKind; id: string }>) => {
			if (!state.draft || state.status !== 'ready' || state.draft.widgets.length >= 30) return;
			const widget = createDashboardWidget(payload.kind, payload.id);
			state.draft.widgets.push({ ...widget, ...vacantPosition(widget, state.draft.widgets) });
			state.message = 'Виджет добавлен. Выберите его содержание в настройках.';
		},
		removeWidget: (state, { payload }: PayloadAction<string>) => {
			if (!state.draft || state.status !== 'ready') return;
			state.draft.widgets = state.draft.widgets.filter((item) => item.id !== payload);
			state.message = 'Виджет убран с главной. Его данные не удалены.';
		},
		updateWidget: (state, { payload }: PayloadAction<DashboardWidget>) => {
			if (!state.draft || state.status !== 'ready' || !state.draft.widgets.some((item) => item.id === payload.id && item.type === payload.type)) return;
			const widget = { ...payload, x: Math.min(payload.x, 12 - payload.w) };
			const widgets = arrangeGrid(state.draft.widgets.map((item) => item.id === widget.id ? widget : item), widget);
			const result = dashboardConfigSchema.safeParse({ ...state.draft, widgets });
			if (result.success) { state.draft = result.data; state.message = 'Настройки виджета применены к черновику.'; }
		},
		moveWidget: (state, { payload }: PayloadAction<{ id: string; x: number; y: number }>) => {
			if (!state.draft || state.status !== 'ready') return;
			const item = state.draft.widgets.find((widget) => widget.id === payload.id);
			if (!item) return;
			const widget = { ...item, x: Math.max(0, Math.min(12 - item.w, payload.x)), y: Math.max(0, Math.min(200, payload.y)) };
			const result = dashboardConfigSchema.safeParse({ ...state.draft, widgets: arrangeGrid(state.draft.widgets, widget) });
			if (result.success) { state.draft = result.data; state.message = 'Виджет перемещён.'; }
		},
		reorderWidget: (state, { payload }: PayloadAction<{ id: string; index: number }>) => {
			if (!state.draft || state.status !== 'ready') return;
			state.draft.widgets = reorderGrid(state.draft.widgets, payload.id, payload.index);
			state.message = 'Порядок виджетов изменён.';
		},
	},
	extraReducers: (builder) => builder.addCase(baseApi.util.resetApiState, () => initialState),
});
export const {
	beginEditing, cancelEditing, resetDraft, addWidget, removeWidget, updateWidget,
	moveWidget, reorderWidget, saveStarted, saveCompleted, saveFailed,
} = dashboardCustomizationSlice.actions;
