import { CreatePlan } from '@features/manage-plans';
import { Page, Tooltip } from '@shared/ui';
import { PlannerBoard } from '@widgets/planner-board';

export const PlannerPage = () => (
	<Page
		title="Планер"
		titleAside={<Tooltip label="Как пользоваться планером" content="Планер помогает сохранять идеи развития бизнеса. Укажите цель, бюджет и дату или оставьте идею без срока. Выберите день, чтобы увидеть планы, а затем план — чтобы изменить его или отметить выполненным." />}
		subtitle="Сохраняйте планы — находите финансирование в нужный момент."
		actions={<CreatePlan />}
	>
		<PlannerBoard />
	</Page>
);
