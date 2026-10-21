import { getApiError } from '@shared/api';
import { useCallback, useEffect, useRef } from 'react';

import { usePatchApplicationMutation } from '../api/applicationApi';
import { type ApplicationPatch, type PatchApplicationResult } from './types';

/** Версия заявки устарела: её изменили в другом окне или вкладке */
export const isVersionConflict = (error: unknown): boolean => getApiError(error)?.code === 'VERSION_CONFLICT';

/**
 * Автосохранение черновика. Сохранения идут строго по очереди, каждое — с актуальной версией в If-Match,
 * поэтому быстрые правки подряд не вызывают ложный VERSION_CONFLICT.
 */
export const useDraftPatch = (applicationId: string, version: number) => {
	const [patch, { isLoading }] = usePatchApplicationMutation();
	const versionRef = useRef(version);
	const queue = useRef<Promise<unknown>>(Promise.resolve());

	useEffect(() => {
		versionRef.current = version;
	}, [version]);

	const save = useCallback((body: ApplicationPatch): Promise<PatchApplicationResult> => {
		const run = queue.current.then(async () => {
			const result = await patch({ id: applicationId, version: versionRef.current, body }).unwrap();
			versionRef.current = result.version;

			return result;
		});
		queue.current = run.catch(() => undefined);

		return run;
	}, [applicationId, patch]);

	return { save, isSaving: isLoading };
};
