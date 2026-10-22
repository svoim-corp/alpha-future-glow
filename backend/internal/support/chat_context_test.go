package support

import "testing"

func TestFinanceContextAndExactAmounts(t *testing.T) {
	previous := &financeContext{Amount: 150000, Term: 12, Goal: "EQUIPMENT", BusinessUse: "unknown"}
	for _, tt := range []struct {
		question string
		amount   int64
		term     int
		product  string
	}{
		{"Рассчитай платёж по экспресс-кредиту с этими параметрами", 150000, 12, "EXPRESS"},
		{"А если 1,5 млн на 2 года по инвестиционному кредиту?", 1500000, 24, "INVEST"},
		{"Хочу 150к на год", 150000, 12, ""},
		{"Хочу 1.25 млн на 18 месяцев", 1250000, 18, ""},
	} {
		p := financeParameters(intentParams{}, tt.question, previous)
		if p.Amount != tt.amount || p.Term != tt.term || p.ProductCode != tt.product || p.Goal != "EQUIPMENT" {
			t.Fatalf("%s: %+v", tt.question, p)
		}
	}
	p := financeParameters(intentParams{}, "Телефоны для сотрудников", previous)
	if p.BusinessUse != "business" || p.Amount != 150000 || p.Term != 12 {
		t.Fatal("не продолжен ответ на уточнение", p)
	}
	p = financeParameters(intentParams{Goal: "REAL_ESTATE"}, "Нужен кредит на помещение", previous)
	if p.Amount != 0 || p.Term != 0 {
		t.Fatal("новая цель получила старые параметры", p)
	}
	p = financeParameters(intentParams{Amount: 2000000, Term: 24, Goal: "EQUIPMENT", BusinessUse: "unknown"}, "Подбери кредит на покупку оборудования за 2 миллиона на 24 месяца", nil)
	if p.BusinessUse != "business" {
		t.Fatal("деловая цель ошибочно требует уточнения личной покупки", p)
	}
	for _, q := range []string{"Покажи мои заявки", "Чем отличается экспресс-кредит?", "Открой планер", "Привет"} {
		if financeContinuation(q) {
			t.Fatal("новая задача принята за ответ", q)
		}
	}
}
