import styles from './Field.module.scss';

interface FieldMessagesProps {
	hint?: string;
	hintId?: string;
	error?: string;
	errorId?: string;
}

export const FieldMessages = ({ hint, hintId, error, errorId }: FieldMessagesProps) => (
	<>
		{hint && <p className={styles.hint} id={hintId}>{hint}</p>}
		{error && <p className={styles.error} id={errorId} role="alert">{error}</p>}
	</>
);
