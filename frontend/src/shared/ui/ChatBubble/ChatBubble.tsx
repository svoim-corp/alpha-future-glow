import { type ReactNode } from 'react';

import styles from './ChatBubble.module.scss';

export interface ChatBubbleProps {
	speaker: 'user' | 'assistant';
	children: ReactNode;
}

export const ChatBubble = ({ speaker, children }: ChatBubbleProps) => (
	<article className={speaker === 'user' ? styles.user : styles.assistant} aria-label={speaker === 'user' ? 'Вы' : 'Ассистент'}>
		<strong className={styles.speaker}>{speaker === 'user' ? 'Вы' : 'Ассистент'}</strong>
		{children}
	</article>
);
