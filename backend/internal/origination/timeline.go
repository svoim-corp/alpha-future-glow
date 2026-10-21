package origination

import (
	"time"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// Чей ход в таймлайне клиента.
const (
	turnClient = "CLIENT"
	turnBank   = "BANK"
)

// Состояние шага таймлайна.
const (
	stepDone    = "DONE"
	stepCurrent = "CURRENT"
	stepPending = "PENDING"
)

// statusText — понятный клиенту текст «что сейчас происходит» и «что дальше» по статусу.
type statusText struct {
	whatsHappening, nextStep, turn string
}

var statusTexts = map[shared.Status]statusText{
	shared.StatusDraft:         {"Заявка ещё не отправлена в банк", "Заполните заявку и отправьте её", turnClient},
	shared.StatusSubmitted:     {"Заявка поступила в банк", "Ничего делать не нужно: банк начинает проверку", turnBank},
	shared.StatusDataCheck:     {"Банк проверяет данные компании по своим источникам и реестрам", "Дождитесь результата проверки", turnBank},
	shared.StatusInfoRequested: {"Банк проверил заявку и просит уточнить данные", "Ответьте на запрос банка ниже", turnClient},
	shared.StatusScoring:       {"Банк оценивает финансовое положение компании", "Обычно это занимает несколько минут", turnBank},
	shared.StatusOnDecision:    {"Заявка на рассмотрении у банка", "Дождитесь решения — мы пришлём уведомление", turnBank},
	shared.StatusApproved:      {"Банк одобрил кредит", "Ознакомьтесь с предложением и примите его", turnClient},
	shared.StatusCounterOffer:  {"Банк готов выдать кредит на других условиях", "Сравните условия и примите или отклоните предложение", turnClient},
	shared.StatusRejected:      {"Банк отказал в кредите", "Посмотрите причину и другие продукты, которые можно оформить", turnClient},
	shared.StatusSigning:       {"Договор готов к подписанию", "Подпишите договор — после этого банк выдаст кредит", turnClient},
	shared.StatusIssued:        {"Кредит выдан", "График платежей — в разделе «Мои кредиты»", turnClient},
	shared.StatusClosed:        {"Заявка закрыта", "Если кредит всё ещё нужен, оформите новую заявку", turnClient},
}

// upcoming — обычный путь заявки дальше текущего статуса. После «На решении» исход неизвестен, поэтому шагов нет.
var upcoming = map[shared.Status][]shared.Status{
	shared.StatusDraft:         {shared.StatusSubmitted, shared.StatusDataCheck, shared.StatusScoring, shared.StatusOnDecision},
	shared.StatusSubmitted:     {shared.StatusDataCheck, shared.StatusScoring, shared.StatusOnDecision},
	shared.StatusDataCheck:     {shared.StatusScoring, shared.StatusOnDecision},
	shared.StatusInfoRequested: {shared.StatusDataCheck, shared.StatusScoring, shared.StatusOnDecision},
	shared.StatusScoring:       {shared.StatusOnDecision},
	shared.StatusApproved:      {shared.StatusSigning, shared.StatusIssued},
	shared.StatusCounterOffer:  {shared.StatusSigning, shared.StatusIssued},
	shared.StatusSigning:       {shared.StatusIssued},
}

type timelineStep struct {
	status shared.Status
	at     *time.Time
	state  string
}

// steps — пройденные статусы из истории по порядку (с повторами, если заявка возвращалась на проверку), текущий
// и ожидаемые. Черновик в таймлайне не показывается: путь клиента в банке начинается с отправки.
func steps(current shared.Status, events []eventRow) []timelineStep {
	var out []timelineStep
	for _, ev := range events {
		if to := shared.Status(ev.To); to != shared.StatusDraft {
			at := ev.At.UTC()
			out = append(out, timelineStep{status: to, at: &at, state: stepDone})
		}
	}
	if n := len(out); n > 0 && out[n-1].status == current {
		out[n-1].state = stepCurrent
	}
	for _, s := range upcoming[current] {
		out = append(out, timelineStep{status: s, state: stepPending})
	}
	return out
}
