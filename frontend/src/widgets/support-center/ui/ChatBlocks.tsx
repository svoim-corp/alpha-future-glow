import { AssistantApplicationCard } from '@entities/application';
import { AssistantCompanyCard } from '@entities/company';
import { AssistantLoanCard, AssistantPaymentCard } from '@entities/loan';
import { AssistantPlanCard } from '@entities/plan';
import { AssistantCalculationCard, AssistantProductCard } from '@entities/product';
import { type ChatBlock } from '@entities/support';
import { ConfirmSupportAction } from '@features/confirm-support-action';
import { DeleteDraftButton } from '@features/delete-draft';
import { formatMoney } from '@shared/lib/format';
import { Alert, ButtonLink, Card, KeyValueList } from '@shared/ui';

interface ChatBlocksProps {
	blocks: ChatBlock[];
	conversationId: string;
}

export const ChatBlocks = ({ blocks, conversationId }: ChatBlocksProps) => blocks.map((b, index) => (
	<div key={b.actionId ?? `${b.kind}-${index}`}>
		{b.product && <AssistantProductCard product={b.product} />}
		{b.calculation && <AssistantCalculationCard calculation={b.calculation} />}
		{b.application && <AssistantApplicationCard application={b.application} actions={b.application.canDelete
			? <DeleteDraftButton applicationId={b.application.id} productName={`Заявка на ${formatMoney(b.application.amount)}`} />
			: undefined} />}
		{b.payment && <AssistantPaymentCard payment={b.payment} />}
		{b.plan && <AssistantPlanCard plan={b.plan} />}
		{b.loan && <AssistantLoanCard loan={b.loan} />}
		{b.company && <AssistantCompanyCard company={b.company} />}
		{b.navigation && <ButtonLink to={b.navigation.href}>{b.navigation.label}</ButtonLink>}
		{b.ticket && <Card><h3>{b.ticket.subject}</h3><KeyValueList items={[
			{ label: 'Номер', value: b.ticket.id },
			{ label: 'Статус', value: b.ticket.status === 'OPEN' ? 'Рассматривается' : b.ticket.status === 'ANSWERED' ? 'Есть ответ' : 'Закрыто' },
		]} /><p>{b.ticket.text}</p>{b.ticket.reply && <p>{`Ответ: ${b.ticket.reply}`}</p>}</Card>}
		{b.kind === 'action' && b.actionId && b.action && <ConfirmSupportAction conversationId={conversationId} actionId={b.actionId} action={b.action} />}
		{b.kind === 'result' && !b.application && <Alert tone="success" title={b.title}>{b.resultId}</Alert>}
	</div>
));
