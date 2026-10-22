import {
	createDefaultDashboard, dashboardApi, dashboardConfigSchema, readLegacyDashboardConfig,
	type DashboardConfig, type DashboardScope, type SaveDashboardArgs,
	useGetDashboardQuery, useSaveDashboardMutation,
} from '@entities/dashboard';
import { getApiError, getErrorMessage, getErrorStatus } from '@shared/api';
import { useAppDispatch, useAppSelector } from '@shared/lib/store';
import { useEffect, useRef, useState } from 'react';

import { beginEditing, cancelEditing, saveCompleted, saveFailed, saveStarted } from './slice';

const defaults = createDefaultDashboard();
const saveErrorText = (error: unknown) => {
	if (getErrorStatus(error) === 412) return 'Главную уже изменили в другом окне или на другом устройстве. Ваш черновик остался открыт. Загрузите сохранённую версию, чтобы продолжить настройку.';
	const message = getErrorMessage(error, 'Не удалось сохранить настройку главной. Черновик остался открыт; повторите сохранение.');
	const fields = [...new Set((getApiError(error)?.fields ?? []).map((field) => field.message))];
	return fields.length ? `${message}. ${fields.join('. ')}` : message;
};

export const useDashboardEditor = (scope: DashboardScope) => {
	const dispatch = useAppDispatch();
	const state = useAppSelector((root) => root.dashboardCustomization);
	const matches = state.scope?.userId === scope.userId && state.scope.companyId === scope.companyId;
	const editing = matches && state.draft !== null;
	const query = useGetDashboardQuery(scope, {
		refetchOnMountOrArgChange: true, refetchOnFocus: !editing, refetchOnReconnect: !editing,
	});
	const [save] = useSaveDashboardMutation();
	const attempt = useRef<{ fingerprint: string; args: SaveDashboardArgs } | null>(null);
	const busy = useRef(false);
	const mounted = useRef(true);
	const [legacy] = useState(() => {
		try { return { config: readLegacyDashboardConfig(scope), error: null }; }
		catch (error) { return { config: null, error: error instanceof Error ? error.message : 'Не удалось прочитать настройку устройства.' }; }
	});
	const [importError, setImportError] = useState<string | null>(null);
	useEffect(() => { mounted.current = true; return () => { mounted.current = false; }; }, []);
	const document = query.currentData;
	const config = editing ? state.draft ?? defaults : document?.config ?? defaults;
	const saving = matches && state.status === 'saving';
	const ready = editing || (!query.isError && document !== undefined);
	const startEditing = (imported?: DashboardConfig) => {
		if (editing) return !saving;
		if (document === undefined || query.isError || query.isFetching || saving) return false;
		setImportError(null);
		dispatch(beginEditing({ scope, document, config: imported }));
		return true;
	};
	const saveDashboard = async () => {
		if (!editing || !state.draft || saving || busy.current) return;
		const parsed = dashboardConfigSchema.safeParse(state.draft);
		if (!parsed.success) { setImportError('Проверьте размеры и расположение виджетов перед сохранением.'); return; }
		const fingerprint = JSON.stringify([scope, state.revision, parsed.data]);
		let command = attempt.current;
		if (!command || command.fingerprint !== fingerprint) {
			command = { fingerprint, args: { scope, config: parsed.data, expectedRevision: state.revision, idempotencyKey: crypto.randomUUID() } };
			attempt.current = command;
		}
		const { args } = command;
		busy.current = true;
		setImportError(null);
		dispatch(saveStarted(args.idempotencyKey));
		try {
			const saved = await save(args).unwrap();
			if (mounted.current) await dispatch(dashboardApi.util.upsertQueryData('getDashboard', scope, saved));
			dispatch(saveCompleted(args.idempotencyKey));
			attempt.current = null;
			// Идемпотентный повтор может вернуть прежний успех; перечитываем актуальную запись.
			if (mounted.current) void query.refetch();
		}
		catch (error) { dispatch(saveFailed({ requestId: args.idempotencyKey, message: saveErrorText(error) })); }
		finally { busy.current = false; }
	};
	const reload = () => {
		if (saving || busy.current) return;
		dispatch(cancelEditing()); attempt.current = null; setImportError(null);
		void query.refetch();
	};
	const importLegacy = () => {
		if (document !== null) return;
		try {
			const imported = readLegacyDashboardConfig(scope);
			if (imported) startEditing(imported);
			else setImportError('На этом устройстве нет сохранённой настройки для переноса.');
		}
		catch (error) { setImportError(error instanceof Error ? error.message : 'Не удалось прочитать настройку устройства.'); }
	};
	return {
		config, editing, ready,
		status: saving ? 'saving' : !editing && query.isFetching ? 'loading' : ready ? 'ready' : 'error',
		error: (matches ? state.error : null) ?? importError ?? (query.isError ? getErrorMessage(query.error, 'Не удалось загрузить настройку главной. Повторите загрузку.') : null),
		message: matches ? state.message : '',
		hasChanges: editing && JSON.stringify(state.draft) !== JSON.stringify(state.baseline),
		canImport: !editing && document === null && Boolean(legacy.config || legacy.error),
		beginEditing: () => startEditing(), save: saveDashboard, reload, importLegacy,
	};
};
