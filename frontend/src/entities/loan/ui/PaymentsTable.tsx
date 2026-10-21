import { formatDate, formatMoney } from '@shared/lib/format';
import { DataTable } from '@shared/ui';

import { type LoanPayment } from '../model/types';
import { PaymentStatusBadge } from './PaymentStatusBadge';

export const PaymentsTable = ({ payments }: { payments: LoanPayment[] }) => (
	<DataTable
		caption="График платежей"
		rows={payments}
		getRowKey={(payment) => payment.date}
		columns={[
			{ key: 'date', header: 'Дата', render: (payment) => formatDate(payment.date) },
			{ key: 'principal', header: 'Основной долг', align: 'right', render: (p) => formatMoney(p.principal) },
			{ key: 'interest', header: 'Проценты', align: 'right', render: (p) => formatMoney(p.interest) },
			{ key: 'total', header: 'Платёж', align: 'right', render: (p) => formatMoney(p.total) },
			{ key: 'status', header: 'Статус', render: (p) => <PaymentStatusBadge status={p.status} /> },
		]}
	/>
);
