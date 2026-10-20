import { FAQ_STEPS, FaqList, type FaqStep, TicketList } from '@entities/support';
import { CreateTicket } from '@features/create-ticket';
import { Card, SelectField } from '@shared/ui';
import { useState } from 'react';
import { useSearchParams } from 'react-router';

import styles from './SupportCenter.module.scss';

export const SupportCenter = () => {
	const [params] = useSearchParams();
	const [step, setStep] = useState<FaqStep>('APPLICATION');
	const applicationId = params.get('applicationId') ?? undefined;

	return (
		<div className={styles.center}>
			<Card className={styles.card} aria-labelledby="s-faq">
				<h2 id="s-faq" className={styles.title}>Частые вопросы</h2>
				<SelectField
					label="Шаг"
					value={step}
					onChange={(event) => setStep(event.target.value as FaqStep)}
					options={FAQ_STEPS}
				/>
				<FaqList step={step} />
			</Card>
			<Card className={styles.card} aria-labelledby="s-new">
				<h2 id="s-new" className={styles.title}>Написать в поддержку</h2>
				<CreateTicket applicationId={applicationId} />
			</Card>
			<Card className={styles.card} aria-labelledby="s-mine">
				<h2 id="s-mine" className={styles.title}>Мои обращения</h2>
				<TicketList />
			</Card>
		</div>
	);
};
