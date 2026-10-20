import { applicationHref, BANK_STATUS_LABEL, STATUS_VIEW } from './status';
import { type ApplicationStatus } from './types';

const ALL: ApplicationStatus[] = [
	'DRAFT', 'SUBMITTED', 'DATA_CHECK', 'INFO_REQUESTED', 'SCORING', 'ON_DECISION',
	'APPROVED', 'COUNTER_OFFER', 'REJECTED', 'SIGNING', 'ISSUED', 'CLOSED',
];

describe('STATUS_VIEW', () => {
	it.each(ALL)('для статуса %s есть клиентская подпись и тон', (status) => {
		expect(STATUS_VIEW[status].label).not.toBe('');
		expect(BANK_STATUS_LABEL[status]).not.toBe('');
	});

	it('иные условия — предупреждение, а не ошибка; отказ — danger', () => {
		expect(STATUS_VIEW.COUNTER_OFFER.tone).toBe('warning');
		expect(STATUS_VIEW.REJECTED.tone).toBe('danger');
		expect(STATUS_VIEW.ISSUED.tone).toBe('success');
	});

	it('сотрудник видит технические подписи', () => {
		expect(BANK_STATUS_LABEL.SCORING).toBe('Скоринг');
		expect(BANK_STATUS_LABEL.ON_DECISION).toBe('На решении');
	});
});

describe('applicationHref', () => {
	it('черновик ведёт в мастер, поданная заявка — в карточку', () => {
		expect(applicationHref({ id: 'a-1', status: 'DRAFT' })).toBe('/applications/a-1/edit');
		expect(applicationHref({ id: 'a-1', status: 'INFO_REQUESTED' })).toBe('/applications/a-1');
	});
});
