import { FAQ_STEPS, FaqList, type FaqStep, TicketList } from '@entities/support';
import { CreateTicket } from '@features/create-ticket';
import { ROUTES } from '@shared/config';
import { Icon, SearchField, SelectFilter } from '@shared/ui';
import { useState } from 'react';
import { Link, useSearchParams } from 'react-router';

import styles from './SupportCenter.module.scss';

/** Центр поддержки по макету support.html: частые вопросы с поиском слева, обращения к специалисту справа */
export const SupportCenter = () => {
	const [params] = useSearchParams();
	const [step, setStep] = useState<FaqStep>('APPLICATION');
	const [query, setQuery] = useState('');
	const applicationId = params.get('applicationId') ?? undefined;

	return (
		<div className={styles.layout}>
			<section className={styles.faq} aria-labelledby="s-faq">
				<div className={styles.intro}>
					<span className={styles.icon}><Icon name="support" size={20} /></span>
					<h2 id="s-faq">Чем можем помочь?</h2>
					<p>Ответы на частые вопросы — без ожидания.</p>
					<div className={styles.search}>
						<SearchField
							label="Поиск по вопросам"
							placeholder="Например, документы или статус заявки"
							value={query}
							onChange={setQuery}
						/>
						<SelectFilter
							label="Шаг"
							value={step}
							onChange={(event) => setStep(event.target.value as FaqStep)}
						>
							{FAQ_STEPS.map((item) => <option key={item.value} value={item.value}>{item.label}</option>)}
						</SelectFilter>
					</div>
				</div>
				<div className={styles.card}>
					<FaqList step={step} query={query} />
				</div>
				<div className={styles.context}>
					<span className={styles.contextIcon}><Icon name="applications" size={18} /></span>
					<div>
						<strong>Вопрос по конкретной заявке?</strong>
						<p>Укажите её номер в обращении — так проще сохранить контекст.</p>
					</div>
					<Link to={ROUTES.applications} className={styles.link}>Мои заявки <Icon name="arrow" size={14} /></Link>
				</div>
			</section>
			<section className={styles.tickets} aria-labelledby="s-new">
				<div className={styles.ticketsHead}>
					<span className={styles.avatar}><Icon name="support" size={22} /></span>
					<div>
						<h2 id="s-new">Написать в поддержку</h2>
						<p>Ответ появится в списке обращений ниже</p>
					</div>
				</div>
				<div className={styles.form}>
					<CreateTicket applicationId={applicationId} />
				</div>
				<div className={styles.mine}>
					<h3 id="s-mine">Мои обращения</h3>
					<TicketList />
				</div>
			</section>
		</div>
	);
};
