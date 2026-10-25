import { FAQ_STEPS, FaqList, type FaqStep, SupportDetails, TicketList } from '@entities/support';
import { CreateTicket } from '@features/create-ticket';
import { ROUTES } from '@shared/config';
import { Icon, panelId, SearchField, SelectFilter, tabId, Tabs } from '@shared/ui';
import { useState } from 'react';
import { Link, useSearchParams } from 'react-router';

import styles from './SupportCenter.module.scss';
import { SupportChat } from './SupportChat';

export const SupportCenter = () => {
	const [params, setParams] = useSearchParams();
	const requestedTab = params.get('tab');
	const tab = requestedTab === 'faq' || requestedTab === 'tickets' || requestedTab === 'assistant'
		? requestedTab
		: params.has('ticket') ? 'tickets' : params.has('faq') ? 'faq' : 'assistant';
	const [step, setStep] = useState<FaqStep>('APPLICATION');
	const [query, setQuery] = useState('');
	const applicationId = params.get('applicationId') ?? undefined;
	const changeTab = (next: string) => setParams((current) => {
		const updated = new URLSearchParams(current);
		updated.set('tab', next);
		return updated;
	}, { replace: true });

	return (
		<div className={styles.support}>
			<Tabs label="Разделы поддержки" items={[{ id: 'assistant', label: 'Ассистент' }, { id: 'faq', label: 'Частые вопросы' }, { id: 'tickets', label: 'Мои обращения' }]} value={tab} onChange={changeTab} />
			<div role="tabpanel" id={panelId(tab)} aria-labelledby={tabId(tab)}>
				{tab === 'assistant' && <SupportChat key={applicationId ?? 'general'} applicationId={applicationId} />}
				{tab === 'faq' && <section className={styles.faq} aria-labelledby="s-faq">
					<div className={styles.intro}>
						<span className={styles.icon}><Icon name="support" size={20} /></span>
						<h2 id="s-faq">Чем можем помочь?</h2>
						<p>Ответы на частые вопросы — без ожидания.</p>
						<div className={styles.search}>
							<SearchField label="Поиск по вопросам" placeholder="Например, документы или статус заявки" value={query} onChange={setQuery} />
							<SelectFilter label="Шаг" value={step} options={FAQ_STEPS} onChange={(next) => setStep(next as FaqStep)} />
						</div>
					</div>
					<div className={styles.card}><FaqList step={step} query={query} /></div>
					<div className={styles.context}>
						<span className={styles.contextIcon}><Icon name="applications" size={18} /></span>
						<div>
							<strong>Вопрос по конкретной заявке?</strong>
							<p>Укажите её номер в обращении — так проще сохранить контекст.</p>
						</div>
						<Link to={ROUTES.applications} className={styles.link}>Мои заявки <Icon name="arrow" size={14} /></Link>
					</div>
				</section>}
				{tab === 'tickets' && <section className={styles.tickets} aria-labelledby="s-new">
					<div className={styles.ticketsHead}><span className={styles.avatar}><Icon name="support" size={22} /></span><div><h2 id="s-new">Написать в поддержку</h2><p>Ответ появится в списке обращений ниже</p></div></div>
					<div className={styles.form}><CreateTicket applicationId={applicationId} /></div>
					<div className={styles.mine}><h3 id="s-mine">Мои обращения</h3><TicketList /></div>
				</section>}
			</div>
			{params.get('ticket') ? <SupportDetails kind="ticket" id={params.get('ticket') ?? ''} onClose={() => {
				const next = new URLSearchParams(params);
				next.set('tab', tab); next.delete('ticket'); next.delete('faq'); setParams(next, { replace: true });
			}} /> : params.get('faq') ? <SupportDetails kind="faq" id={params.get('faq') ?? ''} onClose={() => {
				const next = new URLSearchParams(params);
				next.set('tab', tab); next.delete('faq'); setParams(next, { replace: true });
			}} /> : null}
		</div>
	);
};
