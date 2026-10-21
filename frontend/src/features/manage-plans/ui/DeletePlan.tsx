import { Button } from '@shared/ui';
import { useState } from 'react';

import { DeletePlanDialog } from './DeletePlanDialog';

export const DeletePlan = ({ id, title }: { id: string; title: string }) => {
	const [open, setOpen] = useState(false);

	return (
		<>
			<Button variant="ghost" compact onClick={() => setOpen(true)}>Удалить план</Button>
			{open && <DeletePlanDialog id={id} title={title} onClose={() => setOpen(false)} />}
		</>
	);
};
