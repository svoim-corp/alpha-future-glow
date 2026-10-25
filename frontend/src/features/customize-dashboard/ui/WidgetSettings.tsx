import { useGetAllMyApplicationsQuery } from '@entities/application';
import { DASHBOARD_DEFINITIONS, dashboardWidgetSchema, type DashboardWidget, type QuickAction } from '@entities/dashboard';
import { useGetLoansQuery } from '@entities/loan';
import { useGetPlansQuery } from '@entities/plan';
import { Alert, Button, CheckboxField, Modal, SelectField } from '@shared/ui';
import { useState } from 'react';

import styles from './Customization.module.scss';

const ACTIONS: { id: QuickAction; label: string }[] = [
	{ id: 'apply', label: 'Подать заявку' }, { id: 'plan', label: 'Открыть планер' },
	{ id: 'credits', label: 'Мои кредиты' }, { id: 'calendar', label: 'Платёжный календарь' }, { id: 'support', label: 'Написать банку' },
];
interface WidgetSettingsProps { widget: DashboardWidget; position: number; total: number; onApply: (widget: DashboardWidget, position: number) => void; onClose: () => void }

export const WidgetSettings = ({ widget, position, total, onApply, onClose }: WidgetSettingsProps) => {
	const [draft, setDraft] = useState<DashboardWidget>(widget);
	const [order, setOrder] = useState(position);
	const [error, setError] = useState('');
	const needsLoans = draft.type === 'loan' || draft.type === 'payment' || draft.type === 'payments';
	const loans = useGetLoansQuery(undefined, { skip: !needsLoans });
	const applications = useGetAllMyApplicationsQuery(undefined, { skip: draft.type !== 'application' });
	const plans = useGetPlansQuery(undefined, { skip: draft.type !== 'planFunding' });
	const definition = DASHBOARD_DEFINITIONS[draft.type];
	const loanId = draft.type === 'loan' || draft.type === 'payment' || draft.type === 'payments' ? draft.settings.loanId : null;
	const loanOptions = [
		{ value: '', label: draft.type === 'loan' ? 'Выберите кредит' : 'Все действующие кредиты' },
		...(loans.data ?? []).map((loan) => ({ value: loan.id, label: `${loan.product.name} · ${loan.id}${loan.status === 'CLOSED' ? ' · закрыт' : ''}` })),
		...(loanId && !loans.data?.some((loan) => loan.id === loanId) ? [{ value: loanId, label: `Сохранённый кредит · ${loanId}` }] : []),
	];
	const quickActions = draft.type === 'quick' ? draft.settings.actions : [];
	const applicationId = draft.type === 'application' ? draft.settings.applicationId : null;
	const planId = draft.type === 'planFunding' ? draft.settings.planId : null;
	const planOptions = [
		{ value: '', label: 'Выберите план' },
		...(plans.data?.items ?? []).map((plan) => ({ value: plan.id, label: `${plan.title} · ${plan.id}` })),
		...(planId && !plans.data?.items.some((plan) => plan.id === planId) ? [{ value: planId, label: `Сохранённый план · ${planId}` }] : []),
	];
	const applicationOptions = [
		{ value: '', label: 'Выберите заявку' },
		...(applications.data ?? []).map((application) => ({ value: application.id, label: `${application.product.name} · ${application.number ?? 'Черновик'} · ${application.id}` })),
		...(applicationId && !applications.data?.some((application) => application.id === applicationId) ? [{ value: applicationId, label: `Сохранённая заявка · ${applicationId}` }] : []),
	];
	const apply = () => {
		const result = dashboardWidgetSchema.safeParse(draft);
		if (result.success) onApply(result.data, order);
		else setError('Проверьте настройки виджета.');
	};
	return <Modal title={`Настроить: ${definition.title}`} size="wide" onClose={onClose}
		footer={<><Button variant="secondary" onClick={onClose}>Отмена</Button><Button onClick={apply}>Применить</Button></>}>
		<div className={styles.settings}>
			<p className={styles.hint}>Оформление закреплено за типом. Изменения сохранятся вместе с настройкой главной.</p>
			<SelectField label="Положение на главной" value={String(order)}
				options={Array.from({ length: total }, (_, index) => ({ value: String(index), label: `${index + 1} из ${total}` }))}
				onChange={(value) => setOrder(Number(value))} />
			<SelectField label="Размер на компьютере" value={`${draft.w},${draft.h}`}
				options={definition.sizes.map(([w, h]) => ({ value: `${w},${h}`, label: `${w} × ${h} клеток` }))}
				onChange={(value) => { const size = definition.sizes.find(([w, h]) => `${w},${h}` === value); if (size) setDraft({ ...draft, w: size[0], h: size[1] }); }}
				hint="На планшете и телефоне виджет перестроится автоматически." />
			{needsLoans && <>
				<SelectField label="Кредит" value={loanId ?? ''} options={loanOptions} disabled={loans.isLoading}
					onChange={(id) => setDraft((current) => {
						if (current.type === 'payments') return { ...current, settings: { ...current.settings, loanId: id || null } };
						if (current.type === 'loan' || current.type === 'payment') return { ...current, settings: { loanId: id || null } };
						return current;
					})} />
				{loans.isError && <Alert tone="danger" title="Не удалось загрузить список кредитов." actions={<Button variant="secondary" onClick={() => void loans.refetch()}>Повторить</Button>} />}
			</>}
			{draft.type === 'planFunding' && <>
				<SelectField label="План" value={planId ?? ''} options={planOptions} disabled={plans.isLoading}
					onChange={(id) => setDraft((current) => current.type === 'planFunding' ? { ...current, settings: { planId: id || null } } : current)} />
				{plans.isError && <Alert tone="danger" title="Не удалось загрузить планы." actions={<Button variant="secondary" onClick={() => void plans.refetch()}>Повторить</Button>} />}
				{(plans.data?.items.length ?? 0) >= 200 && <Alert tone="warning" title="Показаны до 200 планов. Список может быть неполным." />}
			</>}
			{draft.type === 'loans' && <SelectField label="Какие кредиты показывать" value={draft.settings.status}
				options={[{ value: 'ACTIVE', label: 'Действующие' }, { value: 'CLOSED', label: 'Закрытые' }]}
				onChange={(value) => { if (value === 'ACTIVE' || value === 'CLOSED') setDraft((current) => current.type === 'loans' ? { ...current, settings: { status: value } } : current); }} />}
			{draft.type === 'payments' && <SelectField label="Период от сегодняшнего дня" value={String(draft.settings.days)}
				options={[{ value: '7', label: 'Ближайшие 7 дней' }, { value: '30', label: 'Ближайшие 30 дней' }]}
				onChange={(value) => { const days = value === '7' ? 7 : 30; setDraft((current) => current.type === 'payments' ? { ...current, settings: { ...current.settings, days } } : current); }} />}
			{draft.type === 'offers' && <SelectField label="Какие предложения показывать" value={draft.settings.filter}
				options={[{ value: 'current', label: 'Действующие и принятые' }, { value: 'all', label: 'Включая истёкшие' }]}
				onChange={(value) => { if (value === 'current' || value === 'all') setDraft((current) => current.type === 'offers' ? { ...current, settings: { filter: value } } : current); }} />}
			{draft.type === 'notifications' && <SelectField label="Уведомления" value={draft.settings.filter}
				options={[{ value: 'unread', label: 'Непрочитанные' }, { value: 'all', label: 'Последние события' }]}
				onChange={(value) => { if (value === 'unread' || value === 'all') setDraft((current) => current.type === 'notifications' ? { ...current, settings: { filter: value } } : current); }} />}
			{draft.type === 'tickets' && <SelectField label="Обращения" value={draft.settings.filter}
				options={[{ value: 'active', label: 'Открытые и с ответом' }, { value: 'answered', label: 'С ответом банка' }, { value: 'all', label: 'Последние обращения' }]}
				onChange={(value) => { if (value === 'active' || value === 'answered' || value === 'all') setDraft((current) => current.type === 'tickets' ? { ...current, settings: { filter: value } } : current); }} />}
			{draft.type === 'application' && <>
				<SelectField label="Заявка" value={applicationId ?? ''} options={applicationOptions} disabled={applications.isLoading}
					onChange={(id) => setDraft((current) => current.type === 'application' ? { ...current, settings: { applicationId: id || null } } : current)} />
				{applications.isError && <Alert tone="danger" title="Не удалось загрузить заявки." actions={<Button variant="secondary" onClick={() => void applications.refetch()}>Повторить</Button>} />}
			</>}
			{draft.type === 'applications' && <SelectField label="Какие заявки показывать" value={draft.settings.filter}
				options={[{ value: 'all', label: 'Все заявки' }, { value: 'attention', label: 'Требуют внимания' }]}
				onChange={(value) => { if (value === 'all' || value === 'attention') setDraft((current) => current.type === 'applications' ? { ...current, settings: { filter: value } } : current); }} />}
			{draft.type === 'calendar' && <SelectField label="События" value={draft.settings.source}
				options={[{ value: 'all', label: 'Платежи и планы' }, { value: 'payments', label: 'Только платежи' }, { value: 'plans', label: 'Только планы' }]}
				onChange={(value) => { if (value === 'all' || value === 'payments' || value === 'plans') setDraft((current) => current.type === 'calendar' ? { ...current, settings: { source: value } } : current); }} />}
			{draft.type === 'plans' && <SelectField label="Какие планы показывать" value={draft.settings.filter}
				options={[{ value: 'upcoming', label: 'Ближайшие планы' }, { value: 'ideas', label: 'Идеи без даты' }]}
				onChange={(value) => { if (value === 'upcoming' || value === 'ideas') setDraft((current) => current.type === 'plans' ? { ...current, settings: { filter: value } } : current); }} />}
			{draft.type === 'quick' && <fieldset className={styles.actions}><legend>Действия и их порядок</legend>
				{[...quickActions, ...ACTIONS.map((action) => action.id).filter((id) => !quickActions.includes(id))].map((id) => {
					const index = quickActions.indexOf(id);
					return <div className={styles.action} key={id}>
						<CheckboxField label={ACTIONS.find((action) => action.id === id)?.label ?? id} checked={index >= 0}
							disabled={index >= 0 && quickActions.length === 1}
							onChange={(event) => { const checked = event.target.checked; setDraft((current) => current.type === 'quick' ? { ...current, settings: { actions: checked ? [...current.settings.actions, id] : current.settings.actions.filter((action) => action !== id) } } : current); }} />
						{index >= 0 && <Button variant="ghost" compact disabled={index === 0} aria-label={`Переместить выше: ${ACTIONS.find((action) => action.id === id)?.label}`} onClick={() => {
							const actions = [...quickActions]; const previous = actions[index - 1];
							if (previous) { actions[index - 1] = id; actions[index] = previous; setDraft((current) => current.type === 'quick' ? { ...current, settings: { actions } } : current); }
						}}>↑</Button>}
					</div>;
				})}
			</fieldset>}
			{error && <Alert tone="danger" title={error} />}
		</div>
	</Modal>;
};
