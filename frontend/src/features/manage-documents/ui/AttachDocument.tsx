import { useGetMyApplicationsQuery } from '@entities/application';
import { getErrorMessage } from '@shared/api';
import { Alert, AsyncBoundary, Button, Icon, Modal, RadioGroup } from '@shared/ui';
import { useState } from 'react';

import { useAttachDocumentMutation } from '../api/attachApi';

/** «В заявку»: выбрать черновик, к которому прикрепить файл из библиотеки */
export const AttachDocument = ({ id, name }: { id: string; name: string }) => {
	const [attach, { isLoading }] = useAttachDocumentMutation();
	const apps = useGetMyApplicationsQuery(undefined, { skip: false });
	const [open, setOpen] = useState(false);
	const [selected, setSelected] = useState('');
	const [failure, setFailure] = useState<string | null>(null);
	const [done, setDone] = useState(false);
	const drafts = (apps.data?.items ?? []).filter((item) => item.status === 'DRAFT');

	const submit = async () => {
		setFailure(null);

		try {
			await attach({ applicationId: selected, documentId: id }).unwrap();
			setDone(true);
			setOpen(false);
		} catch (error) {
			setFailure(getErrorMessage(error, 'Не удалось прикрепить файл'));
		}
	};

	return (
		<>
			<Button
				variant="ghost"
				compact
				aria-label={`В заявку: ${name}`}
				onClick={() => {
					setDone(false);
					setOpen(true);
				}}
			>
				<Icon name="applications" size={15} /> В заявку
			</Button>
			{done && <Alert tone="success" title="Файл прикреплён к черновику" />}
			{open && (
				<Modal
					title="Прикрепить к заявке"
					onClose={() => setOpen(false)}
					footer={(
						<>
							<Button variant="ghost" onClick={() => setOpen(false)}>Отмена</Button>
							<Button onClick={() => void submit()} loading={isLoading} disabled={!selected}>
								Прикрепить
							</Button>
						</>
					)}
				>
					<AsyncBoundary
						isLoading={apps.isLoading}
						isError={apps.isError}
						isEmpty={drafts.length === 0}
						onRetry={() => void apps.refetch()}
						emptyTitle="Нет черновиков"
						emptyText="Файл можно прикрепить только к незавершённой заявке."
					>
						<RadioGroup
							legend="Черновик заявки"
							value={selected}
							onChange={setSelected}
							options={drafts.map((item) => ({
								value: item.id,
								label: `${item.product.name}, ${item.amount.toLocaleString('ru-RU')} ₽`,
							}))}
						/>
					</AsyncBoundary>
					{failure && <Alert tone="danger" title={failure} />}
				</Modal>
			)}
		</>
	);
};
