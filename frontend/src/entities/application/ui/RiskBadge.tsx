import { Badge, type Tone } from '@shared/ui';

import { type DecisionMode, type RiskBand } from '../model/bank';

const RISK: Record<RiskBand, { label: string; tone: Tone }> = {
	LOW: { label: 'Низкий риск', tone: 'success' },
	MEDIUM: { label: 'Средний риск', tone: 'warning' },
	HIGH: { label: 'Высокий риск', tone: 'danger' },
};

export const RiskBadge = ({ band }: { band: RiskBand | null }) => (
	band ? <Badge tone={RISK[band].tone}>{RISK[band].label}</Badge> : <span>—</span>
);

export const DecisionModeLabel = ({ mode }: { mode: DecisionMode }) => (
	<span>{mode === 'AUTO' ? 'Автоматически' : 'Ручной разбор'}</span>
);
