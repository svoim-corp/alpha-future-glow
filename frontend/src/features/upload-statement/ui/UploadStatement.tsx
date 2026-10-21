import { getApiError, getErrorMessage } from '@shared/api';
import { formatDate, formatMoney } from '@shared/lib/format';
import { Alert, Button, FileUpload, KeyValueList } from '@shared/ui';
import { useState } from 'react';

import { type StatementResult, useUploadStatementMutation } from '../api/statementApi';
import styles from './UploadStatement.module.scss';

const MAX_BYTES = 10 * 1024 * 1024;

interface UploadStatementProps {
	applicationId: string;
	disabled?: boolean;
}

/** Выписка из 1С (1CClientBankExchange, .txt, до 10 МБ) — подтверждает обороты и поднимает лимит */
export const UploadStatement = ({ applicationId, disabled = false }: UploadStatementProps) => {
	const [upload, { isLoading }] = useUploadStatementMutation();
	const [file, setFile] = useState<File | null>(null);
	const [fileError, setFileError] = useState<string | undefined>();
	const [result, setResult] = useState<StatementResult | null>(null);
	const [failure, setFailure] = useState<string | null>(null);

	const send = async () => {
		if (!file) {
			return;
		}

		setFailure(null);

		try {
			setResult(await upload({ applicationId, file }).unwrap());
		} catch (error) {
			setResult(null);
			setFailure(getApiError(error)?.code === 'STATEMENT_FORMAT'
				? 'Не удалось прочитать выписку. Нужен файл из 1С формата 1CClientBankExchange с ИНН вашей компании.'
				: getErrorMessage(error, 'Не удалось загрузить выписку. Повторите попытку.'));
		}
	};

	return (
		<div className={styles.block}>
			<FileUpload
				label="Выписка из 1С"
				accept=".txt"
				hint="Файл .txt в формате 1CClientBankExchange, до 10 МБ"
				maxBytes={MAX_BYTES}
				file={file}
				error={fileError}
				disabled={disabled}
				onSelect={(next) => {
					setFile(next);
					setFileError(undefined);
					setFailure(null);
				}}
				onReject={setFileError}
			/>
			<div>
				<Button onClick={() => void send()} loading={isLoading} disabled={!file || disabled}>
					Загрузить выписку
				</Button>
			</div>
			{failure && <Alert tone="danger" title="Выписка не принята">{failure}</Alert>}
			{result && (
				<>
					<KeyValueList
						items={[
							{ label: 'Период', value: `${formatDate(result.period.from)} — ${formatDate(result.period.to)}` },
							{ label: 'Операций', value: result.operations },
							{ label: 'Средние поступления', value: `${formatMoney(result.avgMonthlyIncome)} в месяц` },
							{ label: 'Новый лимит', value: formatMoney(result.newLimit) },
						]}
					/>
					{result.amountFits
						? <Alert tone="success" title="Запрошенная сумма укладывается в новый лимит" />
						: <Alert tone="warning" title="Запрошенная сумма выше нового лимита">Уменьшите сумму в параметрах заявки.</Alert>}
				</>
			)}
		</div>
	);
};
