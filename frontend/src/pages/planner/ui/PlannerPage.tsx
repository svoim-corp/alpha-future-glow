import { CreatePlan } from '@features/manage-plans';
import { Page } from '@shared/ui';
import { PlannerBoard } from '@widgets/planner-board';

export const PlannerPage = () => (
	<Page
		title="Планер"
		subtitle="Сохраняйте планы — находите финансирование в нужный момент."
		actions={<CreatePlan />}
	>
		<PlannerBoard />
	</Page>
);
