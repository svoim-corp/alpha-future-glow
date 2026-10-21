import { useGetGoalsQuery } from '@entities/product';
import { zodResolver } from '@hookform/resolvers/zod';
import { getErrorMessage } from '@shared/api';
import { Alert, Button, CheckboxField, Field, Modal, RadioGroup, SelectField, TextAreaField } from '@shared/ui';
import { useState } from 'react';
import { Controller, useForm, useWatch } from 'react-hook-form';

import { useCreatePlanMutation } from '../api/plansApi';
import { planSchema, type PlanValues } from '../model/schema';

const MODES = [
	{ value: 'UNSCHEDULED', label: 'Без даты' },
	{ value: 'DAY', label: 'Один день' },
	{ value: 'PERIOD', label: 'Период' },
] as const;

const EMPTY: PlanValues = {
	title: '',
	budget: '',
	goal: '',
	dateMode: 'UNSCHEDULED',
	startDate: '',
	endDate: '',
	note: '',
	email: false,
};

export const CreatePlan = () => {
	const [create, { isLoading }] = useCreatePlanMutation();
	const goals = useGetGoalsQuery();
	const [open, setOpen] = useState(false);
	const [failure, setFailure] = useState<string | null>(null);
	const { register, handleSubmit, reset, control, formState: { errors } } = useForm<PlanValues>({
		resolver: zodResolver(planSchema),
		defaultValues: EMPTY,
	});
	const mode = useWatch({ control, name: 'dateMode' });

	const close = () => {
		setOpen(false);
		setFailure(null);
		reset(EMPTY);
	};

	const onSubmit = handleSubmit(async (values) => {
		setFailure(null);

		try {
			await create({
				title: values.title,
				note: values.note === '' ? null : values.note,
				goal: values.goal === '' ? null : values.goal,
				budgetRub: Number(values.budget),
				dateMode: values.dateMode,
				startDate: values.dateMode === 'UNSCHEDULED' ? null : values.startDate,
				endDate: values.dateMode === 'UNSCHEDULED' ? null : (values.dateMode === 'DAY' ? values.startDate : values.endDate),
				notifications: { inApp: true, email: values.email },
			}).unwrap();
			close();
		} catch (error) {
			setFailure(getErrorMessage(error, 'Не удалось сохранить план. Введённое сохранено — повторите попытку.'));
		}
	});

	return (
		<>
			<Button onClick={() => setOpen(true)}>Новый план</Button>
			{open && (
				<Modal
					title="Новый план"
					onClose={close}
					footer={(
						<>
							<Button variant="ghost" onClick={close}>Отмена</Button>
							<Button onClick={() => void onSubmit()} loading={isLoading}>Сохранить план</Button>
						</>
					)}
				>
					<Field label="Название" error={errors.title?.message} {...register('title')} />
					<Field
						label="Ориентировочный бюджет, ₽"
						inputMode="numeric"
						hint="Это пожелание, а не кредитное обязательство"
						error={errors.budget?.message}
						{...register('budget')}
					/>
					<Controller
						control={control}
						name="goal"
						render={({ field }) => (
							<SelectField
								label="Цель"
								options={[{ value: '', label: 'Не выбрана' }, ...(goals.data ?? []).map((g) => ({ value: g.code, label: g.name }))]}
								value={field.value ?? ''}
								onChange={field.onChange}
							/>
						)}
					/>
					<Controller
						control={control}
						name="dateMode"
						render={({ field }) => (
							<RadioGroup
								legend="Когда"
								options={[...MODES]}
								value={field.value}
								onChange={field.onChange}
							/>
						)}
					/>
					{mode !== 'UNSCHEDULED' && (
						<Field
							type="date"
							label={mode === 'DAY' ? 'Дата' : 'Начало'}
							error={errors.startDate?.message}
							{...register('startDate')}
						/>
					)}
					{mode === 'PERIOD' && (
						<Field type="date" label="Окончание" error={errors.endDate?.message} {...register('endDate')} />
					)}
					<TextAreaField label="Заметка" error={errors.note?.message} {...register('note')} />
					<CheckboxField label="Дополнительно сообщать письмом" {...register('email')} />
					{failure && <Alert tone="danger" title={failure} />}
				</Modal>
			)}
		</>
	);
};
