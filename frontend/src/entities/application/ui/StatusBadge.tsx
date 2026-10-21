import { Badge } from '@shared/ui';

import { STATUS_VIEW } from '../model/status';
import { type ApplicationStatus } from '../model/types';

interface StatusBadgeProps {
	status: ApplicationStatus;
	/** Подпись для кабинета сотрудника вместо клиентской */
	label?: string;
}

export const StatusBadge = ({ status, label }: StatusBadgeProps) => (
	<Badge tone={STATUS_VIEW[status].tone}>{label ?? STATUS_VIEW[status].label}</Badge>
);
