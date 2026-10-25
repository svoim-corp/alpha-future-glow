import { getErrorMessage } from '@shared/api';
import { formatDateTime } from '@shared/lib/format';
import { Alert, Button, Modal } from '@shared/ui';

import { useGetFaqItemQuery, useGetTicketQuery } from '../api/supportApi';
import { FAQ_STEPS } from '../model/types';
import styles from './SupportDetails.module.scss';

interface SupportDetailsProps { kind: 'faq' | 'ticket'; id: string; onClose: () => void }

export const SupportDetails = ({ kind, id, onClose }: SupportDetailsProps) => {
	const faq = useGetFaqItemQuery(id, { skip: kind !== 'faq', refetchOnMountOrArgChange: true });
	const ticket = useGetTicketQuery(id, { skip: kind !== 'ticket', refetchOnMountOrArgChange: true });
	const request = kind === 'faq' ? faq : ticket;
	const question = kind === 'faq' && !faq.isError ? faq.currentData : undefined;
	const message = kind === 'ticket' && !ticket.isError ? ticket.currentData : undefined;
	return <Modal title={question?.question ?? message?.subject ?? (kind === 'faq' ? 'Вопрос и ответ' : 'Обращение')}
		size="large" onClose={onClose}>
		{request.isError ? <Alert tone="danger" title={getErrorMessage(request.error, 'Не удалось открыть результат поиска.')}
			actions={<Button variant="secondary" onClick={() => void request.refetch()}>Повторить</Button>} />
			: !request.currentData ? <p role="status">Загружаем…</p> : <div className={styles.body}>
				{question && <>
					<p>{FAQ_STEPS.find((step) => step.value === question.step)?.label}</p><p>{question.answer}</p>
				</>}
				{message && <>
					<p>{message.id} · {formatDateTime(message.updatedAt)}</p>
					<h3>Ваше обращение</h3><p>{message.text ?? 'Текст обращения отсутствует.'}</p>
					<h3>Ответ банка</h3><p>{message.reply ?? 'Ответа пока нет.'}</p>
				</>}
			</div>}
	</Modal>;
};
