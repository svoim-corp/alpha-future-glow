import { bankDetailsFixture, checksFixture, completenessFixture, historyFixture, scoreFixture } from '@test/fixtures/bank';
import { renderWithProviders } from '@test/render';
import { screen } from '@testing-library/react';

import { ApplicantView, ChecksView, CompletenessView, HistoryView, ScoreView } from './BankReviewViews';

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

	it('история: кто, что и когда', () => {
		renderWithProviders(<HistoryView history={historyFixture} />);

		expect(screen.getByText('Клиент Петрова А. С.: SUBMIT')).toBeInTheDocument();
		expect(screen.getByText('Система: CHECKS_DONE')).toBeInTheDocument();
	});
});
