import { render, screen, within } from '@testing-library/react';

import { DataTable } from './DataTable';

interface Row {
	id: string;
	name: string;
	amount: number;
}

describe('DataTable', () => {
	it('рендерит подписанную таблицу с заголовками и строками', () => {
		render(
			<DataTable<Row>
				caption="Заявки"
				getRowKey={(row) => row.id}
				rows={[{ id: '1', name: 'Пекарня', amount: 500000 }, { id: '2', name: 'Ромашка', amount: 700000 }]}
				columns={[
					{ key: 'name', header: 'Компания', render: (row) => row.name },
					{ key: 'amount', header: 'Сумма', align: 'right', render: (row) => row.amount },
				]}
			/>,
		);

		const table = screen.getByRole('table', { name: 'Заявки' });

		expect(within(table).getByRole('columnheader', { name: 'Сумма' })).toBeInTheDocument();
		expect(within(table).getAllByRole('row')).toHaveLength(3);
		expect(within(table).getByRole('cell', { name: 'Ромашка' })).toBeInTheDocument();
	});
});
