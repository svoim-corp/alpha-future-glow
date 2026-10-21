import { AssistantApplicationCard } from '@entities/application';
import { AssistantPaymentCard } from '@entities/loan';
import { AssistantCalculationCard, AssistantProductCard } from '@entities/product';
import { type ChatBlock } from '@entities/support';
import { ConfirmSupportAction } from '@features/confirm-support-action';
import { Alert } from '@shared/ui';

interface ChatBlocksProps {
	blocks: ChatBlock[];
	conversationId: string;
}

export const ChatBlocks = ({ blocks, conversationId }: ChatBlocksProps) => blocks.map((b, index) => (
	<div key={b.actionId ?? `${b.kind}-${index}`}>
		{b.product && <AssistantProductCard product={b.product} />}
		{b.calculation && <AssistantCalculationCard calculation={b.calculation} />}
		{b.application && <AssistantApplicationCard application={b.application} />}
		{b.payment && <AssistantPaymentCard payment={b.payment} />}
		{b.kind === 'action' && b.actionId && b.action && <ConfirmSupportAction conversationId={conversationId} actionId={b.actionId} action={b.action} />}
		{b.kind === 'result' && !b.application && <Alert tone="success" title={b.title}>{b.resultId}</Alert>}
	</div>
));
