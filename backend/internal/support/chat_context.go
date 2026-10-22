package support

import (
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

// Контекст хранится с ответом, но не попадает в API сообщений или индекс знаний.
type financeContext struct {
	Amount      int64  `json:"amount"`
	Term        int    `json:"term"`
	Goal        string `json:"goal"`
	ProductCode string `json:"productCode"`
	BusinessUse string `json:"businessUse"`
	PendingTool string `json:"pendingTool"`
}

func latestFinance(rows []chatMessage) *financeContext {
	for i := len(rows) - 1; i >= 0; i-- {
		if rows[i].Role == "assistant" && rows[i].Status == "COMPLETED" && rows[i].Payload.V.Finance != nil {
			return rows[i].Payload.V.Finance
		}
	}
	return nil
}

func selectedProduct(question string) string {
	q := strings.ToLower(question)
	for _, p := range []struct{ word, code string }{{"экспресс", "EXPRESS"}, {"инвестицион", "INVEST"}, {"оборотн", "WORKING"}, {"овердрафт", "OVERDRAFT"}} {
		if strings.Contains(q, p.word) {
			return p.code
		}
	}
	return ""
}

func financeContinuation(question string) bool {
	q := strings.ToLower(strings.TrimSpace(question))
	for _, prefix := range []string{"чем ", "что ", "как ", "какие ", "почему "} {
		if strings.HasPrefix(q, prefix) {
			return false
		}
	}
	return strings.Contains(q, "для сотрудник") || strings.Contains(q, "для работы") || strings.Contains(q, "для бизнеса") || strings.Contains(q, "для себя") || strings.Contains(q, "личного") || q == "да" || q == "экспресс" || q == "экспресс-кредит" || q == "инвестиционный" || q == "оборотный" || q == "овердрафт"
}

func financeParameters(in intentParams, question string, previous *financeContext) intentParams {
	q := strings.ToLower(question)
	in.ProductCode = "" // Выбор подтверждается репликой клиента или сохранённым контекстом.
	if previous != nil && (in.Goal == "" || in.Goal == previous.Goal) {
		if in.Amount == 0 {
			in.Amount = previous.Amount
		}
		if in.Term == 0 {
			in.Term = previous.Term
		}
		if in.Goal == "" {
			in.Goal = previous.Goal
		}
		if in.ProductCode == "" {
			in.ProductCode = previous.ProductCode
		}
		if in.BusinessUse == "" || (in.BusinessUse == "unknown" && strings.Contains(q, "эт")) {
			in.BusinessUse = previous.BusinessUse
		}
	}
	if code := selectedProduct(question); code != "" {
		in.ProductCode = code
	}
	for _, purpose := range []string{"для сотрудников", "для работы", "для бизнеса"} {
		if strings.Contains(q, purpose) && !strings.Contains(q, "не "+purpose) {
			in.BusinessUse = "business"
		}
	}
	if strings.Contains(q, "для себя") && !strings.Contains(q, "не для себя") {
		in.BusinessUse = "personal"
	}
	// Числовые сокращения нормализуются точно, без округления денежных значений.
	amount := regexp.MustCompile(`(?i)(\d[\d \x{00a0}]*(?:[.,]\d+)?)\s*(млн|миллион(?:а|ов)?|тыс(?:яч[аиу]?)?\.?|[кk])(?:\s|руб|₽|$)`)
	if found := amount.FindStringSubmatch(q); len(found) > 0 {
		digits := strings.NewReplacer(" ", "", "\u00a0", "", ",", ".").Replace(found[1])
		if value, ok := new(big.Rat).SetString(digits); ok {
			factor := int64(1000)
			if strings.HasPrefix(found[2], "м") {
				factor = 1000000
			}
			value.Mul(value, new(big.Rat).SetInt64(factor))
			if value.IsInt() && value.Num().IsInt64() {
				in.Amount = value.Num().Int64()
			}
		}
	}
	term := regexp.MustCompile(`(\d+)\s*(месяц|мес\.?|год|лет)`)
	if found := term.FindStringSubmatch(q); len(found) > 0 {
		if n, err := strconv.Atoi(found[1]); err == nil {
			if strings.HasPrefix(found[2], "ме") {
				in.Term = n
			} else {
				in.Term = n * 12
			}
		}
	}
	if strings.Contains(q, "на год") {
		in.Term = 12
	}
	if strings.Contains(q, "на полтора года") {
		in.Term = 18
	}
	if in.BusinessUse == "unknown" && !strings.Contains(q, "телефон") && !strings.Contains(q, "смартфон") {
		if previous != nil && (strings.Contains(q, "эт") || financeContinuation(question)) {
			in.BusinessUse = previous.BusinessUse
		} else {
			in.BusinessUse = "business"
		}
	}
	return in
}
