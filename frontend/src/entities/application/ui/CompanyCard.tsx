import { formatDate } from '@shared/lib/format';
import clsx from 'clsx';

import { companyStateLabel, isCompanyActive } from '../model/companyState';
import { type ApplicationCompany } from '../model/types';
import styles from './CompanyCard.module.scss';

interface CompanyCardProps {
	company: ApplicationCompany;
}

/** Карточка компании из реестра ЕГРЮЛ/ЕГРИП: клиент её не правит */
export const CompanyCard = ({ company }: CompanyCardProps) => (
	<div className={styles.card}>
		<p className={styles.source}>{company.type === 'IE' ? 'ЕГРИП' : 'ЕГРЮЛ'}</p>
		<p className={styles.name}>{company.name}</p>
		<p className={styles.meta}>
			{company.type === 'IE' ? 'ОГРНИП' : 'ОГРН'} {company.ogrn} · с {formatDate(company.regDate)} ·{' '}
			<span className={clsx(!isCompanyActive(company.state) && styles.warn)}>
				{companyStateLabel(company.state)}
			</span>
		</p>
	</div>
);
