import { type BusinessPlan, useCreatePlanMutation, useUpdatePlanMutation } from '@entities/plan';
import { zodResolver } from '@hookform/resolvers/zod';
import { getErrorMessage, getFieldErrors } from '@shared/api';
import { Alert, Button, CheckboxField, DateField, Field, Modal, MoneyField, RadioGroup, TextAreaField } from '@shared/ui';
import { useState } from 'react';
import { Controller, useForm, useWatch } from 'react-hook-form';

import { planSchema, type PlanValues } from '../model/schema';
import { changedFields, initialValues } from '../model/values';
import styles from './PlanEditor.module.scss';

interface PlanEditorProps {
	plan?: BusinessPlan;
	initialDate?: string;
	onClose: () => void;
}

const MODES = [
	{ value: 'UNSCHEDULED', label: 'Без даты' },
	{ value: 'DAY', label: 'Один день' },
	{ value: 'PERIOD', label: 'Период' },
];

export const PlanEditor = ({ plan, initialDate, onClose }: PlanEditorProps) => {
	const [create, createState] = useCreatePlanMutation();
	const [update, updateState] = useUpdatePlanMutation();
	const [failure, setFailure] = useState<string | null>(null);
	const [confirmClose, setConfirmClose] = useState(false);
	const { register, control, handleSubmit, setError, formState: { errors, isDirty } } = useForm<PlanValues>({
		resolver: zodResolver(planSchema), defaultValues: initialValues(plan, initialDate),
	});
	const mode = useWatch({ control, name: 'dateMode' });
	const busy = createState.isLoading || updateState.isLoading;

	const close = () => {
		if (!busy) {
			if (isDirty) setConfirmClose(true);
			else onClose();
		}
	};

	const submit = handleSubmit(async (values) => {
		if (busy) return;
		setFailure(null);
		try {
			if (plan) {
				const body = changedFields(values, plan);
				if (Object.keys(body).length > 0) await update({ id: plan.id, body }).unwrap();
			} else {
				await create({
					title: values.title, note: values.note || null, goal: values.goal, budgetRub: Number(values.budget),
					dateMode: values.dateMode,
					startDate: values.dateMode === 'UNSCHEDULED' ? null : values.startDate,
					endDate: values.dateMode === 'UNSCHEDULED' ? null : values.dateMode === 'DAY' ? values.startDate : values.endDate,
					notifications: { inApp: values.inApp, email: false },
				}).unwrap();
			}
			onClose();
		} catch (error) {
			const fields = getFieldErrors(error);
			const mapping: Record<string, keyof PlanValues> = {
				title: 'title', budgetRub: 'budget', goal: 'goal', dateMode: 'dateMode', startDate: 'startDate', endDate: 'endDate', note: 'note',
			};
			for (const [field, message] of Object.entries(fields)) {
				const key = mapping[field];
				if (key) setError(key, { message });
			}
			setFailure(getErrorMessage(error, 'Не удалось сохранить план. Введённые данные сохранены в форме.'));
		}
	});

	return (
		<Modal size="large" title={plan ? 'Редактировать план' : 'Новый план'} onClose={close} footer={confirmClose ? (
			<>
				<span>Не сохранять изменения?</span>
				<Button variant="ghost" onClick={() => setConfirmClose(false)}>Продолжить заполнение</Button>
				<Button variant="danger" onClick={onClose}>Не сохранять</Button>
			</>
		) : (
			<>
				<Button variant="ghost" disabled={busy} onClick={close}>Отмена</Button>
				<Button disabled={busy} loading={busy} onClick={() => void submit()}>Сохранить план</Button>
			</>
		)}>
			<form className={styles.form} onSubmit={(event) => { event.preventDefault(); void submit(); }} noValidate>
				<p className={styles.description}>
					Запишите цель и ориентир бюджета. Это план развития, а не заявка на кредит.
				</p>
				<div className={styles.fullWidth}>
					<Field label="Название" maxLength={200} disabled={busy} error={errors.title?.message} {...register('title')} />
				</div>
				<Controller control={control} name="budget" render={({ field }) => (
					<MoneyField label="Ориентировочный бюджет, ₽" disabled={busy} error={errors.budget?.message}
						value={field.value} onValueChange={field.onChange}
						onBlur={field.onBlur} ref={field.ref} name={field.name} />
				)} />
				<Field label="Цель" placeholder="Например, открыть вторую пекарню" maxLength={64}
					disabled={busy} error={errors.goal?.message} {...register('goal')} />
				<Controller control={control} name="dateMode" render={({ field }) => (
					<RadioGroup legend="Когда" options={MODES.map((option) => ({ ...option, disabled: busy }))} value={field.value} onChange={field.onChange} />
				)} />
				<div className={styles.dates}>
					{mode !== 'UNSCHEDULED' && <Controller control={control} name="startDate" render={({ field }) => (
						<DateField label={mode === 'DAY' ? 'Дата' : 'Начало'} disabled={busy}
							error={errors.startDate?.message} value={field.value} onValueChange={field.onChange}
							onBlur={field.onBlur} ref={field.ref} name={field.name} />
					)} />}
					{mode === 'PERIOD' && <Controller control={control} name="endDate" render={({ field }) => (
						<DateField label="Окончание" disabled={busy} error={errors.endDate?.message}
							value={field.value} onValueChange={field.onChange}
							onBlur={field.onBlur} ref={field.ref} name={field.name} />
					)} />}
				</div>
				<div className={styles.fullWidth}>
					<TextAreaField label="Заметка" maxLength={2000} disabled={busy} error={errors.note?.message} {...register('note')} />
				</div>
				<div className={styles.fullWidth}><CheckboxField label="Уведомления в кабинете" disabled={busy} {...register('inApp')} /></div>
				{failure && <div className={styles.fullWidth}><Alert tone="danger" title={failure} /></div>}
			</form>
		</Modal>
	);
};
