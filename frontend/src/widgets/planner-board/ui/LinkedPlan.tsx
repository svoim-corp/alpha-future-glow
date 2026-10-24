import { useGetPlanQuery } from '@entities/plan';
import { useGetGoalsQuery } from '@entities/product';
import { getErrorMessage } from '@shared/api';
import { Alert, Button, Modal } from '@shared/ui';

import { PlanDetails } from './PlanDetails';

interface LinkedPlanProps { id: string; today: string; onClose: () => void }

export const LinkedPlan = ({ id, today, onClose }: LinkedPlanProps) => {
	const query = useGetPlanQuery(id, { refetchOnMountOrArgChange: true });
	const goals = useGetGoalsQuery();
	if (query.currentData && !query.isError) return <PlanDetails plan={query.currentData} today={today}
		goalName={goals.data?.find((goal) => goal.code === query.currentData?.goal)?.name} onClose={onClose} />;
	return <Modal title="План" size="large" onClose={onClose}>
		{query.isError ? <Alert tone="danger" title={getErrorMessage(query.error, 'Не удалось загрузить план.')}
			actions={<Button variant="secondary" onClick={() => void query.refetch()}>Повторить</Button>} />
			: <p role="status">Загружаем план…</p>}
	</Modal>;
};
