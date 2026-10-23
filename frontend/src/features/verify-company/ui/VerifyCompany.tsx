import { type Application, CompanyCard } from '@entities/application';
import { formatDate } from '@shared/lib/format';
import { Alert, Field } from '@shared/ui';

import styles from './VerifyCompany.module.scss';

interface VerifyCompanyProps {
	application: Application;
	disabled?: boolean;
}

/**
 * Компания заявки: ИНН берётся из аккаунта клиента банка и не редактируется,
 * данные компании и полномочия приходят в заявке. Для представителя по МЧД показываем статус доверенности.
 */
export const VerifyCompany = ({ application }: VerifyCompanyProps) => {
	const isRepresentative = application.authority?.type === 'MCHD';
	const validUntil = application.authority?.validUntil;

	return (
		<div className={styles.block}>
			<Field
				label="ИНН компании или ИП"
				value={application.company?.inn ?? ''}
				hint="Берём из вашего аккаунта, остальные данные подтянули из реестров"
				readOnly
				disabled
			/>
			{application.company && <CompanyCard company={application.company} />}
			{isRepresentative && (
				<Alert
					tone={application.authority?.verified ? 'success' : 'info'}
					title={application.authority?.verified ? 'Полномочия подтверждены' : 'Вы подаёте заявку как представитель'}
				>
					{validUntil ? `Доверенность действует до ${formatDate(validUntil)}.` : 'Заявка подаётся по доверенности.'}
				</Alert>
			)}
		</div>
	);
};
