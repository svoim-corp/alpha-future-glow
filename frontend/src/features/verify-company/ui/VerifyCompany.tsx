import { type Application, CompanyCard } from '@entities/application';
import { zodResolver } from '@hookform/resolvers/zod';
import { getErrorMessage } from '@shared/api';
import { formatDate } from '@shared/lib/format';
import { Alert, Button, Field } from '@shared/ui';
import { useState } from 'react';
import { useForm } from 'react-hook-form';

import { type AuthorityCheck, type CompanyLookup, useCheckAuthorityMutation, useLookupCompanyMutation } from '../api/verifyCompanyApi';
import { innSchema, type InnValues, mchdSchema,type MchdValues } from '../model/schema';
import styles from './VerifyCompany.module.scss';

interface VerifyCompanyProps {
	application: Application;
	disabled?: boolean;
}

/**
 * ИНН → карточка компании из реестра → сверка заявителя с руководителем.
 * Если заявитель не руководитель, здесь же проверяется МЧД: до этого согласие на БКИ и решение недоступны.
 */
export const VerifyCompany = ({ application, disabled = false }: VerifyCompanyProps) => {
	const [lookup, lookupState] = useLookupCompanyMutation();
	const [check, checkState] = useCheckAuthorityMutation();
	const [lookupResult, setLookupResult] = useState<CompanyLookup | null>(null);
	const [authorityResult, setAuthorityResult] = useState<AuthorityCheck | null>(null);
	const [lookupError, setLookupError] = useState<string | null>(null);
	const [authorityError, setAuthorityError] = useState<string | null>(null);

	const innForm = useForm<InnValues>({
		resolver: zodResolver(innSchema),
		defaultValues: { inn: application.company?.inn ?? '' },
	});
	const mchdForm = useForm<MchdValues>({ resolver: zodResolver(mchdSchema), defaultValues: { mchdNumber: '' } });

	const needsMchd = lookupResult?.authorityRequired === 'MCHD' || application.authority?.type === 'MCHD';
	const verified = authorityResult?.verified === true || application.authority?.verified === true;

	const onLookup = innForm.handleSubmit(async ({ inn }) => {
		setLookupError(null);

		try {
			setLookupResult(await lookup({ applicationId: application.id, inn }).unwrap());
		} catch (error) {
			setLookupError(getErrorMessage(error, 'Не удалось получить данные компании. Повторите попытку.'));
		}
	});

	const onCheck = mchdForm.handleSubmit(async ({ mchdNumber }) => {
		setAuthorityError(null);

		try {
			setAuthorityResult(await check({ applicationId: application.id, mchdNumber }).unwrap());
		} catch (error) {
			setAuthorityError(getErrorMessage(error, 'Не удалось проверить доверенность. Повторите попытку.'));
		}
	});

	return (
		<div className={styles.block}>
			<form className={styles.row} onSubmit={(event) => void onLookup(event)} noValidate>
				<Field
					label="ИНН компании или ИП"
					inputMode="numeric"
					disabled={disabled}
					hint="Остальные данные подтянем из реестров"
					error={innForm.formState.errors.inn?.message}
					{...innForm.register('inn')}
				/>
				<Button type="submit" variant="secondary" loading={lookupState.isLoading} disabled={disabled}>
					Найти компанию
				</Button>
			</form>
			{lookupError && <Alert tone="danger" title={lookupError} />}
			{application.company && <CompanyCard company={application.company} />}
			{needsMchd && (
				<div className={styles.block}>
					<Alert tone="info" title="Вы подаёте заявку как представитель">
						{lookupResult
							? `Руководитель по реестру — ${lookupResult.company.head}. Нужна машиночитаемая доверенность (МЧД).`
							: 'Руководитель по реестру другой человек. Нужна машиночитаемая доверенность (МЧД).'}
					</Alert>
					{verified ? (
						<Alert tone="success" title="Полномочия подтверждены">
							{authorityResult?.validUntil
								? `Доверенность действует до ${formatDate(authorityResult.validUntil)}.`
								: 'Параметры заявки и согласие на запрос в БКИ доступны.'}
						</Alert>
					) : (
						<form className={styles.row} onSubmit={(event) => void onCheck(event)} noValidate>
							<Field
								label="Номер машиночитаемой доверенности"
								disabled={disabled}
								hint="Пока полномочия не проверены, параметры, согласие на БКИ и решение недоступны"
								error={mchdForm.formState.errors.mchdNumber?.message}
								{...mchdForm.register('mchdNumber')}
							/>
							<Button type="submit" variant="secondary" loading={checkState.isLoading} disabled={disabled}>
								Проверить доверенность
							</Button>
						</form>
					)}
					{authorityResult && !authorityResult.verified && (
						<Alert tone="warning" title="Доверенность не подошла">
							{`${(authorityResult.reason ?? 'Не удалось подтвердить полномочия').replace(/\.$/u, '')}. Подать заявку может только руководитель.`}
						</Alert>
					)}
					{authorityError && <Alert tone="danger" title={authorityError} />}
				</div>
			)}
		</div>
	);
};
