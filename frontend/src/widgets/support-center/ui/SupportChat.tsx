import { supportApi, useGetChatMessagesQuery, useGetConversationsQuery } from '@entities/support';
import { SendSupportMessage, useCreateConversationMutation } from '@features/send-support-message';
import { getErrorMessage } from '@shared/api';
import { formatDateTime } from '@shared/lib/format';
import { Alert, Button, Card, ChatBubble, SelectField, Spinner } from '@shared/ui';
import { useState } from 'react';
import { Link } from 'react-router';

import { ChatBlocks } from './ChatBlocks';
import styles from './SupportChat.module.scss';

interface SupportChatProps { applicationId?: string }

export const SupportChat = ({ applicationId }: SupportChatProps) => {
	const conversations = useGetConversationsQuery();
	const [selected, setSelected] = useState<string | undefined>();
	const [failure, setFailure] = useState<string | null>(null);
	const [create, creation] = useCreateConversationMutation();
	const matching = conversations.data?.filter((c) => (c.applicationId ?? undefined) === applicationId);
	const id = selected ?? matching?.[0]?.id;
	const cached = supportApi.endpoints.getChatMessages.useQueryState(id ?? '', { skip: !id });
	const messages = useGetChatMessagesQuery(id ?? '', {
		skip: !id,
		pollingInterval: cached.currentData?.some((m) => m.status === 'PENDING') ? 1000 : 0,
		skipPollingIfUnfocused: true,
	});
	const history = messages.currentData;
	const pending = history?.some((m) => m.status === 'PENDING') ?? false;
	const newConversation = async () => {
		setFailure(null);
		try { const c = await create({ applicationId }).unwrap(); setSelected(c.id); }
		catch (error) { setFailure(getErrorMessage(error, 'Не удалось начать диалог')); }
	};
	const loading = conversations.isLoading || (Boolean(id) && !history && !messages.isError);
	const loadFailure = conversations.isError || messages.isError;

	return (
		<div className={styles.chat}>
			<header className={styles.header}>
				<div>
					<h2>Помощь в вашем темпе</h2>
					<p>Объясню условия, покажу данные и помогу подготовить заявку.</p>
				</div>
				<Button variant="secondary" loading={creation.isLoading}
					disabled={conversations.isLoading} onClick={() => void newConversation()}>
					Новый диалог
				</Button>
			</header>
			{applicationId && <Alert title="Вопрос по заявке">Контекст заявки: {applicationId}</Alert>}
			{matching && matching.length > 0 && <SelectField
				label="История диалогов" value={id ?? ''} onChange={setSelected}
				options={matching.map((c) => ({ value: c.id, label: formatDateTime(c.createdAt) }))}
			/>}
			{failure && <Alert tone="danger" title={failure} />}
			{loading && <div role="status"><Spinner /> Загружаем переписку…</div>}
			{loadFailure && <Alert tone="danger" title="Не удалось загрузить переписку" actions={
				<Button variant="secondary" onClick={() => {
					if (id) { void messages.refetch(); }
					void conversations.refetch();
				}}>Повторить</Button>
			}>Частые вопросы и обращения доступны на соседних вкладках.</Alert>}
			{!loading && !loadFailure && <>
				{!history?.length && <Card className={styles.welcome}>
					<h3>С чего начнём?</h3>
					<ul>
						<li>Чем отличается овердрафт от обычного кредита?</li>
						<li>Подбери кредит на оборудование за 2 млн рублей на 24 месяца.</li>
						<li>Покажи мои заявки и ближайшие платежи.</li>
					</ul>
					<p>Это ИИ-ассистент. Условия и цифры проверяйте в карточках; решение по кредиту принимает банк.</p>
				</Card>}
				<div className={styles.messages} role="log" aria-label="Переписка с ассистентом" aria-live="polite" aria-relevant="additions text">
					{history?.map((m) => <ChatBubble key={m.id} speaker={m.role}>
						{m.status === 'PENDING' ? <span role="status"><Spinner /> Ассистент готовит ответ…</span> : <p>{m.text}</p>}
						{m.status === 'FAILED' && <Alert tone="warning" title="Ответ не получен">Отправьте вопрос повторно или откройте частые вопросы.</Alert>}
						{id && <ChatBlocks blocks={m.blocks} conversationId={id} />}
						{m.sources.length > 0 && <details>
							<summary>Материалы по вопросу</summary>
							{m.sources.map((s) => <section key={s.id}>
								<h4>{s.title}</h4><p>{s.excerpt}</p>
								<Link to="/support?tab=faq">Открыть частые вопросы</Link>
							</section>)}
						</details>}
					</ChatBubble>)}
				</div>
			</>}
			<SendSupportMessage
				conversationId={id} applicationId={applicationId}
				busy={pending || loading || loadFailure} onConversationCreated={setSelected}
			/>
		</div>
	);
};
