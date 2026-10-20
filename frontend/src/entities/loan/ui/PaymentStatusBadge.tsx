import { Badge, type Tone } from '@shared/ui';

import { type PaymentStatus } from '../model/types';

const VIEW: Record<PaymentStatus, { label: string; tone: Tone }> = {
	PLANNED: { label: 'Запланирован', tone: 'info' },
	PAID: { label: 'Оплачен', tone: 'success' },
	OVERDUE: { label: 'Просрочен', tone: 'danger' },
};

export const PaymentStatusBadge = ({ status }: { status: PaymentStatus }) => (
	<Badge tone={VIEW[status].tone}>{VIEW[status].label}</Badge>
);
