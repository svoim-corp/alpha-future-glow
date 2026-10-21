package origination

import "github.com/svoim-corp/alpha-future-glow/backend/internal/shared"

const (
	stagePrelim = 1 // параметры, согласие на БКИ и предварительное решение
	stageFull   = 2 // данные банка, поручитель, согласия и отправка
)

// stageState — где заявка в мастере.
type stageState struct {
	stage       int
	prelim      *shared.PrelimView // действующий результат этапа 1; nil — нет или устарел
	invalidated bool               // этап 1 пройден, но с тех пор изменились продукт, сумма или срок
}

// stageOf определяет этап по результату этапа 1 (Scoring). Результат устаревает, если его посчитали для других
// продукта, суммы или срока. Этап 2 открывается по лимиту, если сумма в лимите или клиент уже выбрал, что делать
// с суммой сверх лимита. Поданная заявка — всегда на этапе 2.
func stageOf(a *appRow, p *shared.PrelimView) stageState {
	if a.Status != string(shared.StatusDraft) {
		return stageState{stage: stageFull, prelim: p}
	}
	switch {
	case p == nil:
		return stageState{stage: stagePrelim}
	case p.ProductCode != a.ProductCode || p.Amount != a.Amount || p.Term != a.Term:
		return stageState{stage: stagePrelim, invalidated: true}
	case p.Outcome == shared.PrelimLimit && (p.AmountFits || p.Choice != nil):
		return stageState{stage: stageFull, prelim: p}
	}
	return stageState{stage: stagePrelim, prelim: p}
}
