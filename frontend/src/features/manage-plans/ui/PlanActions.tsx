import { type BusinessPlan, PLAN_STATE_LABELS, useUpdatePlanMutation } from '@entities/plan';
import { getErrorMessage } from '@shared/api';
import { Alert, Button, SelectField } from '@shared/ui';
import { type ReactNode, useState } from 'react';

import { DeletePlanDialog } from './DeletePlanDialog';
import { PlanEditor } from './PlanEditor';

interface PlanActionsProps {
	plan: BusinessPlan;
	renderView?: (actions: ReactNode) => ReactNode;
}

export const PlanActions = ({ plan, renderView }: PlanActionsProps) => {
	const [dialog, setDialog] = useState<'edit' | 'delete' | null>(null);
	const [update, { isLoading }] = useUpdatePlanMutation();
	const [failure, setFailure] = useState<string | null>(null);

	const changeState = async (value: string) => {
		const states: BusinessPlan['state'][] = ['ACTIVE', 'PAUSED', 'DONE', 'ARCHIVED'];
		const state = states.find((key) => key === value);
		if (!state) return;
		setFailure(null);
		try {
			await update({ id: plan.id, body: { state } }).unwrap();
		} catch (error) {
			setFailure(getErrorMessage(error, 'Не удалось изменить состояние плана'));
		}
	};

	if (dialog === 'edit') return <PlanEditor plan={plan} onClose={() => setDialog(null)} />;
	if (dialog === 'delete') {
		return <DeletePlanDialog id={plan.id} title={plan.title} onClose={() => setDialog(null)} />;
	}

	const actions = (
		<>
			<SelectField label="Состояние плана" value={plan.state} disabled={isLoading} options={Object.entries(PLAN_STATE_LABELS).map(([value, label]) => ({ value, label }))} onChange={(value) => void changeState(value)} />
			{failure && <Alert tone="danger" title={failure} />}
			<Button variant="secondary" disabled={isLoading} onClick={() => setDialog('edit')}>Редактировать план</Button>
			<Button variant="ghost" compact onClick={() => setDialog('delete')}>Удалить план</Button>
		</>
	);

	return renderView ? renderView(actions) : actions;
};
