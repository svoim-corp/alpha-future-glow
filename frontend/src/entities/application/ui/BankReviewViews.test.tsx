import { bankDetailsFixture, checksFixture, completenessFixture, historyFixture, scoreFixture } from '@test/fixtures/bank';
import { renderWithProviders } from '@test/render';
import { screen } from '@testing-library/react';

import { ApplicantView, ChecksView, CompletenessView, HistoryView, InfoRequestsView, ScoreView } from './BankReviewViews';

describe('BankReviewViews', () => {
	it('анкета показывает заявителя, полномочия по МЧД, ГИР БО и поручителя', () => {
		renderWithProviders(<ApplicantView data={bankDetailsFixture} />);

		expect(screen.getByText('Петрова Анна Сергеевна')).toBeInTheDocument();
		expect(screen.getByText('Доверенность до 01.03.2027')).toBeInTheDocument();
		expect(screen.getByText('ГИР БО за 2025')).toBeInTheDocument();
		expect(screen.getByText(/Смирнов Олег Викторович, ИНН 771234567890$/)).toBeInTheDocument();
	});

	it('поручитель без согласия на БКИ помечен', () => {
		const data = {
			...bankDetailsFixture,
			guarantor: { ...bankDetailsFixture.guarantor!, bkiConsent: false },
		};
		renderWithProviders(<ApplicantView data={data} />);

		expect(screen.getByText(/нет согласия на БКИ/)).toBeInTheDocument();
	});

	it('полнота: замечания со степенью важности текстом', () => {
		renderWithProviders(<CompletenessView data={completenessFixture} />);

		expect(screen.getByText('Анкета заполнена полностью')).toBeInTheDocument();
		expect(screen.getByText('Внимание')).toBeInTheDocument();
	});

	it('проверки: источник, итог и время запроса', () => {
		renderWithProviders(<ChecksView checks={checksFixture} />);

		expect(screen.getByText('ЕГРЮЛ')).toBeInTheDocument();
		expect(screen.getByText(/1 арбитражное дело/)).toBeInTheDocument();
		expect(screen.getAllByText(/запрос 20\.10\.2026/)).toHaveLength(2);
	});

	it('скоринг: каждому фактору — своё влияние со знаком', () => {
		renderWithProviders(<ScoreView score={scoreFixture} />);

		expect(screen.getByText('+80 баллов')).toBeInTheDocument();
		expect(screen.getByText('-70 баллов')).toBeInTheDocument();
		expect(screen.getByText(/Нужен ручной разбор: Пограничный балл/)).toBeInTheDocument();
		expect(screen.getByText(/Модель предлагает: иные условия/)).toBeInTheDocument();
	});

	it('скоринг: рекомендации одобрить и отказать имеют подпись, а не undefined', () => {
		const { rerender } = renderWithProviders(
			<ScoreView score={{ ...scoreFixture, suggested: { decision: 'APPROVED', amount: 1500000, rate: 20 } }} />,
		);

		expect(screen.getByText(/Модель предлагает: одобрить/)).toBeInTheDocument();

		rerender(<ScoreView score={{ ...scoreFixture, suggested: { decision: 'REJECTED' } }} />);

		expect(screen.getByText(/Модель предлагает: отказать/)).toBeInTheDocument();
	});

	it('внешние проверки по счетам в банке и этапы скоринга в истории подписаны по-русски', () => {
		renderWithProviders(
			<>
				<ChecksView
					checks={[
						{ source: 'BANK', status: 'OK', summary: 'Клиент банка с 2019 года', requestedAt: '2026-10-23T13:56:00Z' },
						{ source: 'BANK_ACCOUNTS', status: 'OK', summary: 'Обороты в норме', requestedAt: '2026-10-23T13:56:00Z' },
					]}
				/>
				<HistoryView
					history={[
						{ at: '2026-10-23T13:55:00Z', actor: { type: 'SYSTEM' }, action: 'SCORING_STARTED', from: 'DATA_CHECK', to: 'SCORING' },
						{ at: '2026-10-23T13:55:00Z', actor: { type: 'SYSTEM' }, action: 'SCORED', from: 'SCORING', to: 'ON_DECISION' },
					]}
				/>
			</>,
		);

		expect(screen.getByText('Отношения с банком')).toBeInTheDocument();
		expect(screen.getByText('Счета и обороты в банке')).toBeInTheDocument();
		expect(screen.getByText('Система: скоринг запущен')).toBeInTheDocument();
		expect(screen.getByText('Система: скоринг завершён')).toBeInTheDocument();
	});

	it('запросы информации: вопрос, ответ клиента и дата, либо «Ответа пока нет»', () => {
		renderWithProviders(
			<InfoRequestsView
				requests={[
					{
						id: 'r-1',
						status: 'ANSWERED',
						comment: 'Обороты расходятся с отчётностью',
						items: [{ key: 'turnover_note', type: 'TEXT', label: 'Почему выручка ниже оборотов' }],
						answers: [{ key: 'turnover_note', value: 'Возврат займа' }],
						createdAt: '2026-10-20T10:00:00Z',
						answeredAt: '2026-10-21T09:00:00Z',
					},
					{
						id: 'r-2',
						status: 'OPEN',
						comment: 'Нужны пояснения',
						items: [{ key: 'purpose', type: 'TEXT', label: 'У кого закупка' }],
						createdAt: '2026-10-22T10:00:00Z',
					},
				]}
			/>,
		);

		expect(screen.getByText('Клиент ответил')).toBeInTheDocument();
		expect(screen.getByText('Почему выручка ниже оборотов')).toBeInTheDocument();
		expect(screen.getByText('Возврат займа')).toBeInTheDocument();
		expect(screen.getByText('Ждём ответ клиента')).toBeInTheDocument();
		expect(screen.getByText('Ответа пока нет')).toBeInTheDocument();
	});

	it('история: кто, что и когда', () => {
		renderWithProviders(<HistoryView history={historyFixture} />);

		expect(screen.getByText('Клиент Петрова А. С.: заявка подана')).toBeInTheDocument();
		expect(screen.getByText('Система: проверки завершены')).toBeInTheDocument();
	});
});
