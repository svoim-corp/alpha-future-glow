import { formatMoney } from '@shared/lib/format';
import { AsyncBoundary, KeyValueList } from '@shared/ui';

import { useGetCompanyProfileQuery } from '../api/applicationApi';

const TAX_REGIME: Record<string, string> = {
	USN_INCOME: 'УСН «Доходы»',
	USN_INCOME_MINUS_EXPENSES: 'УСН «Доходы минус расходы»',
	OSNO: 'Общая система налогообложения',
	PSN: 'Патент',
};

/** Данные из реестров. Клиент их не правит: данные ведёт банк */
export const CompanyProfileView = ({ applicationId }: { applicationId: string }) => {
	const { data, isLoading, isError, refetch } = useGetCompanyProfileQuery(applicationId);

	return (
		<AsyncBoundary
			isLoading={isLoading}
			isError={isError}
			onRetry={() => void refetch()}
			loadingLabel="Загружаем данные компании"
			errorText="Не удалось загрузить данные компании"
		>
			{data && (
				<KeyValueList
					items={[
						{ label: 'Адрес регистрации', value: data.regAddress },
						...(data.actualAddress ? [{ label: 'Фактический адрес', value: data.actualAddress }] : []),
						{ label: 'Основной вид деятельности', value: `${data.okved.code} ${data.okved.name}` },
						{ label: 'Руководитель', value: data.head.fullName },
						{
							label: 'Учредители',
							value: data.founders.map((founder) => `${founder.fullName} — ${founder.share}%`).join(', '),
						},
						{ label: 'Налоговый режим', value: TAX_REGIME[data.taxRegime] ?? data.taxRegime },
						...(data.turnover
							? [{
								label: 'Средние поступления на счета в банке',
								value: `${formatMoney(data.turnover.avgMonthlyIncome)} в месяц`,
							}]
							: []),
						...(data.girbo?.accounted ? [{ label: 'Бухгалтерская отчётность', value: 'Учтена' }] : []),
					]}
				/>
			)}
		</AsyncBoundary>
	);
};
