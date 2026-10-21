import { Button } from '@shared/ui';
import { useState } from 'react';

import { PlanEditor } from './PlanEditor';

interface CreatePlanProps {
	date?: string;
	label?: string;
}

export const CreatePlan = ({ date, label = 'Новый план' }: CreatePlanProps) => {
	const [open, setOpen] = useState(false);

	return (
		<>
			<Button onClick={() => setOpen(true)}>{label}</Button>
			{open && <PlanEditor initialDate={date} onClose={() => setOpen(false)} />}
		</>
	);
};
