import { formatMoney } from '@shared/lib/format';
import { AsyncBoundary, KeyValueList } from '@shared/ui';

import { useGetCompanyProfileQuery } from '../api/applicationApi';

const TAX_REGIME: Record<string, string> = {
	USN_INCOME: 'УСН «Доходы»',
	USN_INCOME_MINUS_EXPENSES: 'УСН «Доходы минус расходы»',
	OSNO: 'Общая система налогообложения',
	PSN: 'Патент',
};

/** Данные из реестров. Клиент их не правит: устаревшее обновляется через налоговую */
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
						{ label: 'Основной вид деятельности', value: `${data.okved.code} ${data.okved.name}` },
						{ label: 'Руководитель', value: data.head },
						{
							label: 'Учредители',
							value: data.founders.map((founder) => `${founder.name} — ${founder.share}%`).join(', '),
						},
						{ label: 'Налоговый режим', value: TAX_REGIME[data.taxRegime] ?? data.taxRegime },
						{ label: 'Клиент банка', value: data.isBankClient ? 'Да' : 'Нет' },
						...(data.internalTurnover
							? [{
								label: 'Средние поступления в банке',
								value: `${formatMoney(data.internalTurnover.avgMonthlyIncome)} в месяц`,
							}]
							: []),
						...(data.girbo?.accounted ? [{ label: 'Бухгалтерская отчётность', value: 'Учтена' }] : []),
					]}
				/>
			)}
		</AsyncBoundary>
	);
};
