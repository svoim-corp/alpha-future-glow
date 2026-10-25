import { AssistantPlanCard } from '@entities/plan';
import { type ChatAction } from '@entities/support';
import { getErrorMessage } from '@shared/api';
import { formatMoney, formatTerm } from '@shared/lib/format';
import { Alert, Button, ButtonLink, Card, KeyValueList } from '@shared/ui';
import { useState } from 'react';

import { useConfirmSupportActionMutation } from '../api/actionApi';

interface ConfirmSupportActionProps {
	conversationId: string;
	actionId: string;
	action: ChatAction;
}

export const ConfirmSupportAction = ({ conversationId, actionId, action }: ConfirmSupportActionProps) => {
	const [confirm, { data, isLoading }] = useConfirmSupportActionMutation();
	const [failure, setFailure] = useState<string | null>(null);
	const execute = async () => {
		setFailure(null);
		try { await confirm({ conversationId, actionId }).unwrap(); }
		catch (error) { setFailure(getErrorMessage(error, 'Не удалось выполнить действие. Повторите подтверждение.')); }
	};

	return (
		<Card>
			<h3>{action.kind === 'draft' ? 'Создать черновик заявки?' : action.kind === 'plan' ? 'Сохранить план?' : 'Отправить обращение?'}</h3>
			{action.draft && <KeyValueList items={[
				{ label: 'Продукт', value: action.productName ?? action.draft.productCode },
				{ label: 'Сумма', value: formatMoney(action.draft.amount) },
				{ label: 'Срок', value: formatTerm(action.draft.term) },
				{ label: 'Цель', value: action.goalLabel ?? action.draft.goal },
			]} />}
			{action.kind === 'ticket' && <><strong>{action.subject}</strong><p>{action.text}</p></>}
			{action.planPreview && <AssistantPlanCard plan={action.planPreview} />}
			{failure && <Alert tone="danger" title={failure} />}
			{data ? <Alert tone="success" title={data.title}>{data.application
				? <ButtonLink to={`/applications/${encodeURIComponent(data.application.id)}/edit`}>Открыть заявку</ButtonLink>
				: data.plan ? <ButtonLink to="/planner">Открыть сохранённый план в планере</ButtonLink>
					: `Номер обращения: ${data.resultId ?? ''}`}</Alert>
				: <Button loading={isLoading} onClick={() => void execute()}>Подтвердить</Button>}
		</Card>
	);
};
